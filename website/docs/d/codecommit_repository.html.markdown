---
subcategory: "CodeCommit"
layout: "aws"
page_title: "AWS: aws_codecommit_repository"
description: |-
  Provides details about CodeCommit Repository.
---

# Data Source: aws_codecommit_repository

The CodeCommit Repository data source allows the ARN, Repository ID, Repository URL for HTTP and Repository URL for SSH to be retrieved for an CodeCommit repository.

## Example Usage

```terraform
data "aws_codecommit_repository" "test" {
  repository_name = "MyTestRepository"
}
```

## Argument Reference

This data source supports the following arguments:

* `repository_name` - (Required) Name for the repository. This needs to be less than 100 characters.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `repository_id` - ID of the repository.
* `kms_key_id` - The ID of the encryption key.
* `arn` - ARN of the repository.
* `clone_url_http` - URL to use for cloning the repository over HTTPS.
* `clone_url_ssh` - URL to use for cloning the repository over SSH.
