FROM golang:alpine3.8 as builder

RUN apk --update upgrade
RUN apk --no-cache --no-progress add make git gcc musl-dev

WORKDIR /build
COPY . .
RUN go build .

FROM node:10-alpine
RUN apk update && apk add --no-cache --virtual ca-certificates
COPY --from=builder /build/eslint-action /usr/bin/eslint-action

LABEL version="1.0.0"
LABEL repository="https://github.com/rkusa/eslint-action"
LABEL homepage="https://github.com/rkusa/eslint-action"
LABEL maintainer="Markus Ast <m@rkusa.st>"

LABEL com.github.actions.name="ESLint"
LABEL com.github.actions.description="Execute ESLint and add issue annotations"
LABEL com.github.actions.icon="octagon"
LABEL com.github.actions.color="#463fd4"

ENV ESLINT_CMD ./node_modules/.bin/eslint
COPY entrypoint.sh /entrypoint.sh

ENTRYPOINT ["/entrypoint.sh"]
CMD [""]