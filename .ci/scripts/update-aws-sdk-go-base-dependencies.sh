#!/bin/sh
# Copyright IBM Corp. 2014, 2025
# SPDX-License-Identifier: MPL-2.0

# Update aws-sdk-go-base dependencies.
go get github.com/hashicorp/aws-sdk-go-base/v2 && go mod tidy
go get github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2 && go mod tidy
git add --update && git commit --message "Update aws-sdk-go-base dependencies."
