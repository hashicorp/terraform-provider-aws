#!/bin/sh

# Update all AWS SDK for Go v2 dependencies.
declare -a StringArray=(`grep github.com/aws/aws-sdk-go-v2 go.mod | grep -v indirect | cut -f2 | cut -d ' ' -f1`)
for val in "${StringArray[@]}"; do
  go get $val && go mod tidy
  git add --update && git commit --message "go get $val."
done

# Update AWS SDK for Go v1 dependency.
go get github.com/aws/aws-sdk-go && go mod tidy
git add --update && git commit --message "go get github.com/aws/aws-sdk-go."
