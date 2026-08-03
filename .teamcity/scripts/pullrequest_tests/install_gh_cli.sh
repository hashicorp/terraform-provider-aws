#!/usr/bin/env bash
# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

version=$(curl -fsSL -H "Authorization: Bearer ${GH_TOKEN}" https://api.github.com/repos/cli/cli/releases/latest | grep '"tag_name"' | sed -E 's/.*"v([^"]+)".*/\1/')
mkdir -p tools && wget -O gh.tar.gz "https://github.com/cli/cli/releases/download/v${version}/gh_${version}_linux_amd64.tar.gz" && tar -xzf gh.tar.gz && mv "gh_${version}_linux_amd64/bin/gh" tools/gh
