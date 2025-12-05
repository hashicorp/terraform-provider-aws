#!/bin/bash
set -euo pipefail

# Script to update copyright headers from HashiCorp to IBM
# Usage: ./update-copyright.sh [--dry-run]

DRY_RUN=false
if [[ "${1:-}" == "--dry-run" ]]; then
	DRY_RUN=true
	echo "DRY RUN MODE - No files will be modified"
fi

# Find all Go files with HashiCorp copyright (including malformed ones)
echo "Finding files with HashiCorp copyright headers..."
FILES=$(rg -l "^// Copyright.*HashiCorp" --type go)
FILE_COUNT=$(echo "$FILES" | wc -l | tr -d ' ')

echo "Found $FILE_COUNT files to update"

# Check for variations
echo ""
echo "Copyright format variations found:"
rg "^// Copyright.*HashiCorp" --type go | cut -d: -f2 | sort -u

if [[ "$DRY_RUN" == "true" ]]; then
	echo ""
	echo "Sample files that would be updated:"
	echo "$FILES" | head -10
	echo ""
	echo "Run without --dry-run to perform the update"
	exit 0
fi

# Perform the replacement
echo ""
echo "Updating copyright headers..."
UPDATED=0
echo "$FILES" | while IFS= read -r file; do
	# Handle standard format
	perl -i -pe 's|^// Copyright \(c\) HashiCorp, Inc\.$|// Copyright IBM Corp. 2014, 2025|' "$file"
	
	# Handle malformed formats (anything with HashiCorp in copyright line)
	perl -i -pe 's|^// Copyright.*HashiCorp.*$|// Copyright IBM Corp. 2014, 2025|' "$file"
	
	UPDATED=$((UPDATED + 1))
	if [[ $((UPDATED % 500)) -eq 0 ]]; then
		echo "  Updated $UPDATED files..."
	fi
done

echo "✓ Updated $FILE_COUNT files"
echo ""
echo "Verification:"
echo "  Old format remaining: $(rg -c '^// Copyright.*HashiCorp' --type go 2>/dev/null | wc -l | tr -d ' ') files"
echo "  New format present: $(rg -c '^// Copyright IBM Corp\. 2014, 2025' --type go 2>/dev/null | wc -l | tr -d ' ') files"
echo ""
echo "Next steps:"
echo "1. Review changes: git diff | head -100"
echo "2. Check specific file: head -5 main.go"
echo "3. Update .copywrite.hcl: mv .copywrite.hcl.new .copywrite.hcl"
echo "4. Run: make fmt"
echo "5. Verify: rg '^// Copyright.*HashiCorp' --type go"
echo "6. Commit changes"
