#!/bin/bash
set -euo pipefail

# Fixes copyright headers by replacing HashiCorp format with IBM format
# and adds missing headers

FIXED=0
ADDED=0

# Explicit exemptions (same as copyright-check.sh)
EXEMPT_FILES=(
	"./internal/service/odb/cloud_vm_clusters_data_source_test.go"
	"./internal/service/eks/token.go"
)

is_exempt() {
	local file="$1"
	for exempt in "${EXEMPT_FILES[@]}"; do
		[[ "$file" == "$exempt" ]] && return 0
	done
	return 1
}

while IFS= read -r file; do
	# Skip exempt files
	is_exempt "$file" && continue
	
	FIRST_LINE=$(head -1 "$file")
	
	# Skip generated files
	[[ "$FIRST_LINE" =~ ^//\ Code\ generated ]] && continue
	
	# Fix HashiCorp headers
	if [[ "$FIRST_LINE" =~ ^//\ Copyright.*HashiCorp ]]; then
		perl -i -pe 's|^// Copyright.*HashiCorp.*$|// Copyright IBM Corp. 2014, 2025|' "$file"
		echo "Fixed: $file"
		FIXED=$((FIXED + 1))
		continue
	fi
	
	# Fix commented-out headers
	if [[ "$FIRST_LINE" =~ ^//\ //\ Copyright ]]; then
		perl -i -pe 's|^// // Copyright.*$|// Copyright IBM Corp. 2014, 2025|' "$file"
		perl -i -pe 's|^// // SPDX-License-Identifier: MPL-2.0$|// SPDX-License-Identifier: MPL-2.0|' "$file"
		echo "Fixed: $file"
		FIXED=$((FIXED + 1))
		continue
	fi
	
	# Add missing header
	if [[ ! "$FIRST_LINE" =~ ^//\ Copyright\ IBM\ Corp ]]; then
		# Create temp file with header
		{
			echo "// Copyright IBM Corp. 2014, 2025"
			echo "// SPDX-License-Identifier: MPL-2.0"
			echo ""
			cat "$file"
		} > "$file.tmp"
		mv "$file.tmp" "$file"
		echo "Added: $file"
		ADDED=$((ADDED + 1))
	fi
done < <(find . -name "*.go" -type f ! -path "./vendor/*" ! -path "./.ci/*" ! -path "./.github/*" ! -path "./.teamcity/*" ! -path "./.release/*")

if [[ $FIXED -eq 0 && $ADDED -eq 0 ]]; then
	echo "✓ No files needed fixing"
else
	echo ""
	[[ $FIXED -gt 0 ]] && echo "✓ Fixed $FIXED files"
	[[ $ADDED -gt 0 ]] && echo "✓ Added headers to $ADDED files"
fi
