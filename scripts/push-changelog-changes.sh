#!/bin/bash

set -o errexit
set -o nounset

EMAIL=$1
UNAME=$2
MSG=$3

if [[ `git status --porcelain` ]]; then
    git config --local user.email $1
    git config --local user.name $2
    git add CHANGELOG.md
    git commit -m "$3" 
    git push
fi
