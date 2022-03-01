#!/usr/bin/env bash

set -euo pipefail

go test ./internal/sweep -v -tags=sweep -sweep="%SWEEPER_REGIONS%" -sweep-allow-failures -timeout=4h
