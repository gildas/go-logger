#!/usr/bin/env bash

#nodemon --verbose --delay 5 --watch . --ignore bin/ --ignore log/ --ignore tmp/ --ignore '*.md' --ignore go.mod --ignore go.sum  --ext go --exec "go test -v . -run 'TestInternalLoggerSuite' || exit 1"
nodemon --verbose --delay 5 --watch . --ignore .git/ --ignore bin/ --ignore log/ --ignore tmp/ --ignore '*.md' --ignore go.mod --ignore go.sum  --ext go --exec "go test -v . || exit 1"
