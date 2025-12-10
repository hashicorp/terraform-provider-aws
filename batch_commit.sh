#!/bin/bash

BATCH_SIZE=100
BATCH_NUM=1

# Get all changed files
git status --porcelain | cut -c4- > /tmp/changed_files.txt
TOTAL_FILES=$(wc -l < /tmp/changed_files.txt)

echo "Total files to commit: $TOTAL_FILES"
echo "Batch size: $BATCH_SIZE"

while IFS= read -r file; do
    echo "$file" >> "/tmp/batch_$BATCH_NUM.txt"
    
    # Check if we've reached batch size or end of files
    CURRENT_BATCH_SIZE=$(wc -l < "/tmp/batch_$BATCH_NUM.txt")
    
    if [ "$CURRENT_BATCH_SIZE" -eq "$BATCH_SIZE" ]; then
        echo "Committing batch $BATCH_NUM ($BATCH_SIZE files)..."
        
        # Add files from current batch
        while IFS= read -r batch_file; do
            git add "$batch_file"
        done < "/tmp/batch_$BATCH_NUM.txt"
        
        # Commit the batch
        git commit -m "Batch $BATCH_NUM: $(head -1 /tmp/batch_$BATCH_NUM.txt | xargs dirname | sort -u | head -3 | tr '\n' ' ')..."
        
        # Clean up and prepare for next batch
        rm "/tmp/batch_$BATCH_NUM.txt"
        BATCH_NUM=$((BATCH_NUM + 1))
    fi
done < /tmp/changed_files.txt

# Handle remaining files in final batch
if [ -f "/tmp/batch_$BATCH_NUM.txt" ]; then
    FINAL_BATCH_SIZE=$(wc -l < "/tmp/batch_$BATCH_NUM.txt")
    echo "Committing final batch $BATCH_NUM ($FINAL_BATCH_SIZE files)..."
    
    while IFS= read -r batch_file; do
        git add "$batch_file"
    done < "/tmp/batch_$BATCH_NUM.txt"
    
    git commit -m "Batch $BATCH_NUM (final): $(head -1 /tmp/batch_$BATCH_NUM.txt | xargs dirname | sort -u | head -3 | tr '\n' ' ')..."
    rm "/tmp/batch_$BATCH_NUM.txt"
fi

# Clean up
rm /tmp/changed_files.txt

echo "All batches committed successfully!"
