#!/bin/sh

# Update Terraform DevEx dependencies.
go get github.com/hashicorp/terraform-plugin-framework && go mod tidy
go get github.com/hashicorp/terraform-plugin-framework-jsontypes && go mod tidy
go get github.com/hashicorp/terraform-plugin-framework-timeouts && go mod tidy
go get github.com/hashicorp/terraform-plugin-framework-timetypes && go mod tidy
go get github.com/hashicorp/terraform-plugin-framework-validators && go mod tidy
go get github.com/hashicorp/terraform-plugin-sdk/v2 && go mod tidy
go get github.com/hashicorp/terraform-plugin-testing && go mod tidy
git add --update && git commit --message "Update Terraform DevEx dependencies."
