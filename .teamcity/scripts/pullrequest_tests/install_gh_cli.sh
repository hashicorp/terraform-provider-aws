#!/usr/bin/env bash
# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

version=$(curl -fsSL \
  ${GH_TOKEN:+-H "Authorization: Bearer ${GH_TOKEN}"} \
  https://api.github.com/repos/cli/cli/releases/latest \
  | grep -oP '"tag_name":\s*"v\K[^"]+')

if [[ -z "${version}" ]]; then
  version="2.97.0"
  echo "WARN: failed to resolve gh CLI version from GitHub API, falling back to ${version}" >&2
fi

mkdir -p tools \
  && wget -O gh.tar.gz "https://github.com/cli/cli/releases/download/v${version}/gh_${version}_linux_amd64.tar.gz" \
  && tar -xzf gh.tar.gz \
  && mv "gh_${version}_linux_amd64/bin/gh" tools/gh
