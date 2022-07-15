#!/usr/bin/env bash

set -euo pipefail

pushd "$GOENV_ROOT"
printf '\nUpdating goenv to %s...\n' "${GOENV_TOOL_VERSION}"
git pull origin "${GOENV_TOOL_VERSION}"
popd

goenv install -s "$(goenv local)" && goenv rehash
