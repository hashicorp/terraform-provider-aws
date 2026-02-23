#!/usr/bin/env bash
# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

set -euo pipefail

if [ $# -ne 1 ]; then
    echo "Usage: $0 <path-to-changelog-file>"
    echo "Example: $0 .changelog/44302.txt"
    exit 1
fi

CHANGELOG_FILE="$1"

if [ ! -f "$CHANGELOG_FILE" ]; then
    echo "Error: File not found: $CHANGELOG_FILE"
    exit 1
fi

# Validate file format
if ! grep -q '```release-note:' "$CHANGELOG_FILE"; then
    echo "Error: File does not contain go-changelog format entries"
    exit 1
fi

# Extract PR number from filename
PR_NUMBER=$(basename "$CHANGELOG_FILE" .txt)

if ! [[ "$PR_NUMBER" =~ ^[0-9]+$ ]]; then
    echo "Error: Could not extract valid PR number from filename"
    exit 1
fi

ENTRY_COUNT=0

# Parse the file and extract entries
while IFS= read -r line; do
    if [[ "$line" =~ ^\`\`\`release-note:(.+)$ ]]; then
        # Start of a new entry
        TYPE="${BASH_REMATCH[1]}"
        BODY=""

        # Read until closing ```
        while IFS= read -r content_line; do
            if [[ "$content_line" == '```' ]]; then
                break
            fi
            if [ -n "$BODY" ]; then
                BODY="${BODY}"$'\n'"${content_line}"
            else
                BODY="${content_line}"
            fi
        done

        # Validate body is not empty
        if [ -z "$BODY" ]; then
            echo "Error: Empty body for entry type '$TYPE'"
            exit 1
        fi

        # Map go-changelog type to Changie kind
        case "$TYPE" in
            new-resource)
                KIND="feature"
                FEATURE_TYPE="New Resource"
                ;;
            new-data-source)
                KIND="feature"
                FEATURE_TYPE="New Data Source"
                ;;
            new-ephemeral-resource)
                KIND="feature"
                FEATURE_TYPE="New Ephemeral Resource"
                ;;
            new-function)
                KIND="feature"
                FEATURE_TYPE="New Function"
                ;;
            new-list-resource)
                KIND="feature"
                FEATURE_TYPE="New List Resource"
                ;;
            enhancement)
                KIND="enhancement"
                FEATURE_TYPE=""
                ;;
            bug)
                KIND="bug"
                FEATURE_TYPE=""
                ;;
            note)
                KIND="note"
                FEATURE_TYPE=""
                ;;
            breaking-change)
                KIND="breaking-change"
                FEATURE_TYPE=""
                ;;
            *)
                echo "Error: Unknown entry type '$TYPE'"
                exit 1
                ;;
        esac

        # Create Changie fragment
        if [ "$KIND" = "feature" ]; then
            # For features, extract the resource/data source name from body
            NAME=$(echo "$BODY" | head -n1)
            changie new --kind "$KIND" --custom "NewType=$FEATURE_TYPE" --custom "Name=$NAME" --custom "PullRequest=$PR_NUMBER"
        else
            # For other kinds (breaking-change, note, enhancement, bug)
            # Extract Impact (resource/data-source prefix) and Body (description) separately
            if [[ "$BODY" =~ ^(resource/|data-source/|ephemeral-resource/|function/|list-resource/|provider)([^:]*):\ (.+)$ ]]; then
                # Has Impact prefix (e.g., "resource/aws_example: Description")
                IMPACT="${BASH_REMATCH[1]}${BASH_REMATCH[2]}"
                DESCRIPTION="${BASH_REMATCH[3]}"
                changie new --kind "$KIND" --custom "Impact=$IMPACT" --custom "Body=$DESCRIPTION" --custom "PullRequest=$PR_NUMBER"
            elif [[ "$BODY" =~ ^(provider):\ (.+)$ ]]; then
                # Provider-level change without resource prefix (e.g., "provider: Description")
                IMPACT="${BASH_REMATCH[1]}"
                DESCRIPTION="${BASH_REMATCH[2]}"
                changie new --kind "$KIND" --custom "Impact=$IMPACT" --custom "Body=$DESCRIPTION" --custom "PullRequest=$PR_NUMBER"
            else
                # No Impact prefix - use full body as description with empty Impact
                # This shouldn't happen in well-formed entries, but handle gracefully
                echo "Warning: Entry does not follow expected format (missing Impact prefix): $BODY"
                changie new --kind "$KIND" --custom "Body=$BODY" --custom "PullRequest=$PR_NUMBER"
            fi
        fi

        ENTRY_COUNT=$((ENTRY_COUNT + 1))
        
        # Add 1 second delay to ensure unique timestamps for Changie files
        # Changie uses second-level precision in filenames (kind-YYYYMMDD-HHMMSS.yaml)
        # Without this, multiple entries of the same kind can overwrite each other
        sleep 1
    fi
done < "$CHANGELOG_FILE"

if [ "$ENTRY_COUNT" -eq 0 ]; then
    echo "Error: No valid entries found in $CHANGELOG_FILE"
    exit 1
fi

# Delete the original file
rm "$CHANGELOG_FILE"

echo "Successfully converted $ENTRY_COUNT entry/entries and deleted: $CHANGELOG_FILE"
