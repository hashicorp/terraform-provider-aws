---
subcategory: "ECR (Elastic Container Registry)"
layout: "aws"
page_title: "AWS: aws_ecr_repository_policy"
description: |-
  Lists ECR (Elastic Container Registry) Repository Policy resources.
---

# List Resource: aws_ecr_repository_policy

Lists ECR (Elastic Container Registry) Repository Policy resources.
Only repositories with an attached repository policy are returned.

## Example Usage

```terraform
list "aws_ecr_repository_policy" "example" {
  provider = aws
}
```

## Argument Reference

This list resource supports the following arguments:

* `region` - (Optional) Region to query. Defaults to provider region.
