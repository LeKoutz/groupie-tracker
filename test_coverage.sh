#!/usr/bin/env bash
go test -v -coverprofile=cover.out ./...
go tool cover -html=cover.out -o coverage.html
go tool cover -func=cover.out
