#!/usr/bin/env bash

set -euo pipefail

pushd "$GOENV_ROOT"
# Make sure we're using the main `goenv`
if ! git remote | grep -q syndbg; then
  printf '\nInstalling syndbg/goenv\n'
  git remote add -f syndbg https://github.com/syndbg/goenv.git
fi
printf '\nUpdating goenv to %s...\n' "${GOENV_TOOL_VERSION}"
git reset --hard syndbg/"${GOENV_TOOL_VERSION}"
popd

goenv install --skip-existing "$(goenv local)" && goenv rehash
