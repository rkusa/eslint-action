#!/bin/sh

set -e

sh -c "$ESLINT_CMD $* --format json . | /usr/bin/eslint-action"