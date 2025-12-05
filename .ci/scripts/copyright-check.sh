#!/bin/bash
set -euo pipefail

# Drop-in replacement for "copywrite headers"
# Checks that all Go files have the IBM copyright header

EXPECTED="// Copyright IBM Corp. 2014, 2025"
MISSING=0

while IFS= read -r file; do
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
