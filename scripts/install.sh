#!/bin/bash

PROJECT_PATH="$1"
CUSTOM_PATH="$2"

if [[ -z $CUSTOM_PATH ]]; then
    echo "Custom path was not provided. Trying to install into GOBIN..."
else
    go build -o "$CUSTOM_PATH/stt" "$PROJECT_PATH"
    exit
fi

if [[ -z $GOBIN ]]; then
    echo "GOBIN is empty... installing into /usr/local/bin/"
else
    go build -o "$GOBIN/stt" "$PROJECT_PATH"
    exit
fi

sudo go build -o /usr/local/bin/stt "$PROJECT_PATH"
