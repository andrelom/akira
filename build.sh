#!/bin/bash

# Set the app name.
APP_NAME="akira"

# Set the output directory.
OUTPUT_DIR="bin"

# Delete the output directory if it exists.
rm -rf $OUTPUT_DIR

# Create the output directory.
mkdir -p $OUTPUT_DIR

# Compile for MacOS.
GOOS=darwin GOARCH=arm64 go build -o $OUTPUT_DIR/darwin/$APP_NAME

# Compile for Linux.
GOOS=linux GOARCH=amd64 go build -o $OUTPUT_DIR/linux/$APP_NAME

# Compile for Windows.
GOOS=windows GOARCH=amd64 go build -o $OUTPUT_DIR/windows/$APP_NAME\.exe
