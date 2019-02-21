package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/google/go-github/v24/github"
	"github.com/ldez/ghactions"
)

type Message struct {
	RuleId   string
	Severity int
	Message  string
	Line     int
}

type File struct {
	FilePath     string
	Messages     []*Message
	ErrorCount   int
	WarningCount int
}

func main() {
	var report []*File
	err := json.NewDecoder(os.Stdin).Decode(&report)
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()
	action := ghactions.NewAction(ctx)

	action.OnPush(func(client *github.Client, event *github.PushEvent) error {
		return handlePush(ctx, client, event, report)
	})

	if err := action.Run(); err != nil {
		log.Fatal(err)
	}
}

func handlePush(ctx context.Context, client *github.Client, event *github.PushEvent, report []*File) error {
	head := os.Getenv(ghactions.GithubSha)
	owner, repoName := ghactions.GetRepoInfo()

	// find the action's checkrun
	checkName := os.Getenv(ghactions.GithubAction)
	result, _, err := client.Checks.ListCheckRunsForRef(ctx, owner, repoName, head, &github.ListCheckRunsOptions{
		CheckName: github.String(checkName),
		Status:    github.String("in_progress"),
	})
	if err != nil {
		return err
	}

	if len(result.CheckRuns) == 0 {
		return fmt.Errorf("Unable to find check run for action: %s", checkName)
	}
	checkRun := result.CheckRuns[0]

	// add annotations for lint issues
	workspacePath := os.Getenv(ghactions.GithubWorkspace) + "/"
	var annotations []*github.CheckRunAnnotation
	errorCount := 0
	warningCount := 0
	for _, r := range report {
		path := strings.TrimPrefix(r.FilePath, workspacePath)

		errorCount += r.ErrorCount
		warningCount += r.WarningCount
		for _, m := range r.Messages {
			var level *string
			switch m.Severity {
			case 1:
				level = github.String("warning")
			case 2:
				level = github.String("failure")
			}

			annotations = append(annotations, &github.CheckRunAnnotation{
				Path:            github.String(path),
				StartLine:       github.Int(m.Line),
				EndLine:         github.Int(m.Line),
				AnnotationLevel: level,
				Title:           github.String(m.RuleId),
				Message:         github.String(m.Message),
			})
		}
	}

	var summary string
	if errorCount == 0 {
		summary = "No problems found"
	} else {
		summary = fmt.Sprintf(
			"%d problems (%d errors, %d warnings)",
			errorCount+warningCount,
			errorCount,
			warningCount,
		)
	}

	// add annotations in #50 chunks
	for i := 0; i < len(annotations); i += 50 {
		end := i + 50

		if end > len(annotations) {
			end = len(annotations)
		}

		_, _, err = client.Checks.UpdateCheckRun(ctx, owner, repoName, checkRun.GetID(), github.UpdateCheckRunOptions{
			Name:    checkName,
			HeadSHA: github.String(head),
			Output: &github.CheckRunOutput{
				Title:       github.String("Result"),
				Summary:     github.String(summary),
				Annotations: annotations[i:end],
			},
		})
		if err != nil {
			return err
		}
	}

	if errorCount > 0 {
		return fmt.Errorf(summary)
	} else {
		return nil
	}
}
