#!/bin/bash
batch_size=100
batch_num=1

git status --porcelain | cut -c4- | while read -r -d $'\n' -a files; do
    for ((i=0; i<${#files[@]}; i+=batch_size)); do
        batch_files=("${files[@]:i:batch_size}")
        git add "${batch_files[@]}"
        git commit -m "Update copyright headers (batch $batch_num)"
        echo "Committed batch $batch_num with ${#batch_files[@]} files"
        ((batch_num++))
    done
done
