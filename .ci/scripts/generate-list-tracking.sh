#!/usr/bin/env bash
set -euo pipefail

total_resources=0
list_implemented=0
tmp_table=$(mktemp)
trap 'rm -f "$tmp_table"' EXIT

while read -r file; do
    if grep -qE "FrameworkResource|SDKResource" "$file"; then
        total_resources=$((total_resources + 1))
        
        if grep -qE "ArnIdentity|IdentityAttribute" "$file"; then
            has_identity="✅"
        else
            has_identity="❌"
        fi

        list_file="${file%.go}_list.go"
        if [[ -f "$list_file" ]]; then
            has_list="✅"
            list_implemented=$((list_implemented + 1))
        else
            has_list="❌"
        fi

        display_name="${file#internal/service/}"
        echo "| \`$display_name\` | $has_identity | $has_list |" >> "$tmp_table"
    fi
done < <(find internal/service -type f -name "*.go" ! -name "*_test.go" ! -name "*_list.go" ! -name "*_gen.go" | sort)

if [ "$total_resources" -gt 0 ]; then
    pct=$(( list_implemented * 100 / total_resources ))
else
    pct=0
fi

filled=$(( pct / 5 ))
empty=$(( 20 - filled ))
bar=""
for ((i=0; i<filled; i++)); do bar+="▓"; done
for ((i=0; i<empty; i++)); do bar+="░"; done

# Output the dynamic content
cat << EOF

### Implementation Progress:

\`[${bar}] ${pct}% (${list_implemented}/${total_resources})\`

| Resource File | Resource Identity | List |
|---|:---:|:---:|
EOF

if [ -s "$tmp_table" ]; then
    cat "$tmp_table"
fi

echo ""
echo "*Last updated: $(date -u +'%Y-%m-%d %H:%M:%S UTC')*"
