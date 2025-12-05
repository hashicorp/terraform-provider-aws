#!/bin/bash
set -euo pipefail

BATCH_SIZE=100
DRY_RUN=false

if [[ "${1:-}" == "--dry-run" ]]; then
	DRY_RUN=true
	echo "DRY RUN MODE"
fi

echo "Finding files..."
TOTAL=$(rg -l "^// Copyright.*HashiCorp" --type go | wc -l | tr -d ' ')
BATCHES=$(( (TOTAL + BATCH_SIZE - 1) / BATCH_SIZE ))

echo "Total files: $TOTAL"
echo "Batch size: $BATCH_SIZE"
echo "Total commits: $BATCHES"

if [[ "$DRY_RUN" == "true" ]]; then
	exit 0
fi

# Create temp file list
TMPFILE=$(mktemp)
rg -l "^// Copyright.*HashiCorp" --type go > "$TMPFILE"

BATCH_NUM=1
COUNT=0
START=1

while IFS= read -r file; do
	perl -i -pe 's|^// Copyright.*HashiCorp.*$|// Copyright IBM Corp. 2014, 2025|' "$file"
	COUNT=$((COUNT + 1))
	
	if [[ $((COUNT % BATCH_SIZE)) -eq 0 ]]; then
		git add -A
		git commit -m "Update copyright headers (batch $BATCH_NUM/$BATCHES)

Files: $START-$COUNT of $TOTAL"
		echo "✓ Batch $BATCH_NUM/$BATCHES committed"
		BATCH_NUM=$((BATCH_NUM + 1))
		START=$((COUNT + 1))
	fi
done < "$TMPFILE"

# Commit remaining
if git diff --quiet; then
	echo "No remaining changes"
else
	git add -A
	git commit -m "Update copyright headers (batch $BATCH_NUM/$BATCHES)

Files: $START-$COUNT of $TOTAL"
	echo "✓ Final batch committed"
fi

rm "$TMPFILE"
echo ""
echo "Complete. Migration finished."
