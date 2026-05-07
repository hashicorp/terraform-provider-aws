#!/usr/bin/env bash
# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

# Sort changelog entries alphabetically within each kind section.
#
# Usage:
#     ./sort-changelog-entries.sh <changelog-file>
#     
# Example:
#     ./sort-changelog-entries.sh .changes/6.x/6.26.0.md

set -euo pipefail

if [[ $# -ne 1 ]]; then
    echo "Usage: $0 <changelog-file>" >&2
    exit 1
fi

CHANGELOG_FILE="$1"

if [[ ! -f "$CHANGELOG_FILE" ]]; then
    echo "Error: File not found: $CHANGELOG_FILE" >&2
    exit 1
fi

# Create temporary files
TEMP_FILE=$(mktemp)
TEMP_ENTRIES=$(mktemp)
trap "rm -f $TEMP_FILE $TEMP_ENTRIES" EXIT

# Process the file line by line
in_section=false

while IFS= read -r line; do
    # Check if this is a section header (all caps ending with colon)
    if [[ "$line" =~ ^[A-Z][A-Z\ ]+:$ ]]; then
        # If we have accumulated entries, sort and output them
        if [[ -s "$TEMP_ENTRIES" ]]; then
            sort -f "$TEMP_ENTRIES" >> "$TEMP_FILE"
            > "$TEMP_ENTRIES"  # Clear the temp file
        fi
        
        # Output the header
        echo "$line" >> "$TEMP_FILE"
        in_section=true
        
    # Check if this is an entry line (starts with "* ")
    elif [[ "$line" =~ ^\*\  ]] && [[ "$in_section" == true ]]; then
        # Accumulate entry for sorting
        echo "$line" >> "$TEMP_ENTRIES"
        
    # Any other line
    else
        # If we have accumulated entries, sort and output them
        if [[ -s "$TEMP_ENTRIES" ]]; then
            sort -f "$TEMP_ENTRIES" >> "$TEMP_FILE"
            > "$TEMP_ENTRIES"  # Clear the temp file
            in_section=false
        fi
        
        # Output the line as-is
        echo "$line" >> "$TEMP_FILE"
    fi
done < "$CHANGELOG_FILE"

# Handle any remaining entries at end of file
if [[ -s "$TEMP_ENTRIES" ]]; then
    sort -f "$TEMP_ENTRIES" >> "$TEMP_FILE"
fi

# Replace original file with sorted version
mv "$TEMP_FILE" "$CHANGELOG_FILE"

echo "✓ Sorted entries in $CHANGELOG_FILE"
