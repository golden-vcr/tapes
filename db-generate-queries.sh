#!/usr/bin/env bash
set -e

SQLC_VERSION=v1.20.0
go run "github.com/sqlc-dev/sqlc/cmd/sqlc@$SQLC_VERSION" generate
