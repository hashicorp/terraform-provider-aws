#!/usr/bin/env bash
# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

set -euo pipefail

# Update all AWS SDK for Go v2 dependencies.
StringArray=()
while IFS= read -r line; do
  StringArray+=("$line")
done < <(grep github.com/aws/aws-sdk-go-v2 go.mod | grep -v indirect | cut -f2 | cut -d ' ' -f1)
for val in "${StringArray[@]}"; do
  go get $val && go mod tidy
  git add --update && git commit --message "go get $val."
done
