#!/usr/bin/env bash
# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

set -euo pipefail

echo ""
echo "### Resources with List Support Implemented"
echo ""
echo "| Resource Name | Registry Documentation |"
echo "|---|---|"

while read -r file; do
    if grep -qE "FrameworkResource|SDKResource" "$file"; then
        list_file="${file%.go}_list.go"
        if [[ -f "$list_file" ]]; then
            # Extract resource name
            resource_name=$(grep -oE '@(SDKResource|FrameworkResource)\("[^"]+"' "$file" | head -n 1 | cut -d'"' -f2)
            
            if [[ -n "$resource_name" ]]; then
                # Remove leading 'aws_' for the doc link
                doc_path="${resource_name#aws_}"
                # Link to the list resources documentation path
                link="[Link](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/list-resources/${doc_path})"
            else
                resource_name="Unknown"
                link="Unknown"
            fi
            
            echo "| \`${resource_name}\` | $link |"
        fi
    fi
done < <(find internal/service -type f -name "*.go" ! -name "*_test.go" ! -name "*_list.go" ! -name "*_gen.go" | sort)

echo ""
echo "*Last updated: $(date -u +'%Y-%m-%d %H:%M:%S UTC')*"
