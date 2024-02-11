#!/bin/bash

# BUILD

# Get Go version
GO_VERSION=$(go version | awk '{print $3}')

# Get the build date
BUILD_DATE=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

go build -o melodix -ldflags "-X github.com/keshon/dice-roller/internal/version.BuildDate=$BUILD_DATE -X github.com/keshon/dice-roller/internal/version.GoVersion=$GO_VERSION" cmd/main.go && ./dice-roller
