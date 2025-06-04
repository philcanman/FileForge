#!/bin/bash

# Specify the name of the Go file to build
GO_FILE="FileForge.go"

# Specify the name of the application
APP_NAME="FileForge"

# Create a bin directory if it doesn't exist
mkdir -p bin

# Function to build the application for Windows
build_windows() {
    echo "Building for Windows..."
    GOOS=windows GOARCH=amd64 go build -o "bin/${APP_NAME}_windows.exe" "$GO_FILE"
    echo "Build successful for Windows."
}

# Function to build the application for Linux
build_linux() {
    echo "Building for Linux..."
    GOOS=linux GOARCH=amd64 go build -o "bin/${APP_NAME}_linux" "$GO_FILE"
    echo "Build successful for Linux."
}

# Main script
echo "Building Go application for Windows and Linux 64-bit..."

# Run build functions
build_windows
build_linux

echo "All builds completed successfully."
