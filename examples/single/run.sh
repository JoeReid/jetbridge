#!/bin/sh

set -e

# Create a binding for the lambda function (using the repo cli version)
go run ../../cmd/cli/main.go binding create \
    --lambda "$(cat .lambda-arn)" \
    --stream "MESSAGES" \
    --subject "MESSAGES.sent"

# Publish some messages to the stream
nats pub MESSAGES.sent '{"sent_at": "2009-11-10T23:00:00Z", "sent_by": "alice", "message": "hello world"}'
