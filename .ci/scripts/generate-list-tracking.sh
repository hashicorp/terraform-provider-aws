#!/usr/bin/env bash
# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

set -euo pipefail

tmp_dir=$(mktemp -d)
trap 'rm -rf "$tmp_dir"' EXIT

global_total=0
global_implemented=0

while read -r file; do
    if grep -qE "FrameworkResource|SDKResource" "$file"; then
        global_total=$((global_total + 1))
        
        if grep -qE "ArnIdentity|IdentityAttribute" "$file"; then
            has_identity="✅"
        else
            has_identity="❌"
        fi

        list_file="${file%.go}_list.go"
        if [[ -f "$list_file" ]]; then
            has_list="✅"
            global_implemented=$((global_implemented + 1))
            implemented=1
        else
            has_list="❌"
            implemented=0
        fi

        # Clean up the file path for display
        display_name="${file#internal/service/}"
        
        # Extract service name (first directory after internal/service/)
        service_name=$(echo "$display_name" | cut -d'/' -f1)

        # Update service counters
        total_file="$tmp_dir/$service_name.total"
        impl_file="$tmp_dir/$service_name.impl"
        table_file="$tmp_dir/$service_name.table"

        echo 1 >> "$total_file"
        if [ "$implemented" -eq 1 ]; then
            echo 1 >> "$impl_file"
        fi

        # Save the markdown table row
        echo "| \`$display_name\` | $has_identity | $has_list |" >> "$table_file"
    fi
done < <(find internal/service -type f -name "*.go" ! -name "*_test.go" ! -name "*_list.go" ! -name "*_gen.go" | sort)

# Function to print progress bar
print_progress() {
    local impl=$1
    local total=$2
    local prefix=$3
    
    if [ "$total" -gt 0 ]; then
        local pct=$(( impl * 100 / total ))
    else
        local pct=0
    fi

    local filled=$(( pct * 40 / 100 ))
    local empty=$(( 40 - filled ))
    local bar=""
    for ((i=0; i<filled; i++)); do bar+="▓"; done
    for ((i=0; i<empty; i++)); do bar+="░"; done

    echo "${prefix}\`[${bar}] ${pct}% (${impl}/${total})\`"
}

echo ""
echo "### Implementation Progress:"
echo ""
print_progress "$global_implemented" "$global_total" "Overall: "
echo ""
echo "Expand a service below to see service-level progress."
echo ""

# Iterate over services alphabetically
for table in $(ls -1 "$tmp_dir"/*.table 2>/dev/null | sort); do
    service_name=$(basename "$table" .table)
    
    svc_total=$(wc -l < "$tmp_dir/$service_name.total" | xargs)
    if [ -f "$tmp_dir/$service_name.impl" ]; then
        svc_impl=$(wc -l < "$tmp_dir/$service_name.impl" | xargs)
    else
        svc_impl=0
    fi

    if [ "$svc_total" -gt 0 ]; then
        svc_pct=$(( svc_impl * 100 / svc_total ))
    else
        svc_pct=0
    fi

    svc_text="${svc_pct}% (${svc_impl}/${svc_total})"
    if [ "$svc_impl" -eq "$svc_total" ] && [ "$svc_total" -gt 0 ]; then
        svc_text="${svc_text} ✅"
    elif [ "$svc_impl" -gt 0 ] && [ "$svc_impl" -lt "$svc_total" ]; then
        svc_text="${svc_text} 🟡"
    fi

    echo "<details><summary><code>${service_name}</code> ${svc_text}</summary><br>"
    echo ""
    echo "| Resource File | Resource Identity | List |"
    echo "|---|:---:|:---:|"
    cat "$table"
    echo ""
    echo "</details>"
done

echo ""
echo "*Last updated: $(date -u +'%Y-%m-%d %H:%M:%S UTC')*"
