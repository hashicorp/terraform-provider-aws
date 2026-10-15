---
subcategory: "ECR (Elastic Container Registry)"
layout: "aws"
page_title: "AWS: aws_ecr_repository"
description: |-
  Lists ECR repositories.
---

# List Resource: aws_ecr_repository

Lists ECR repositories in a region.

## Example Usage

```terraform
list "aws_ecr_repository" "example" {
  provider = aws

  config {}
}
```

## Argument Reference

This list resource supports the following arguments:

* `region` - (Optional) Region to query. Defaults to provider region.
