#!/bin/bash
set -euo pipefail

EXPECTED="// Copyright IBM Corp. 2014, 2025"
MISSING=0

# Explicit exemptions for third-party copyrights
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
	
	if [[ ! "$FIRST_LINE" =~ ^//\ Copyright\ IBM\ Corp\.\ 2014,\ 2025$ ]]; then
		echo "Missing or incorrect copyright header: $file"
		MISSING=$((MISSING + 1))
	fi
done < <(find . -name "*.go" -type f ! -path "./vendor/*" ! -path "./.ci/*" ! -path "./.github/*" ! -path "./.teamcity/*" ! -path "./.release/*")

if [[ $MISSING -gt 0 ]]; then
	echo ""
	echo "Error: $MISSING files missing correct copyright header"
	echo "Expected: $EXPECTED"
	exit 1
fi

echo "✓ All files have correct copyright headers"
