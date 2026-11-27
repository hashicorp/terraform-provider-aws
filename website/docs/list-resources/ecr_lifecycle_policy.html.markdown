---
subcategory: "ECR (Elastic Container Registry)"
layout: "aws"
page_title: "AWS: aws_ecr_lifecycle_policy"
description: |-
  Lists ECR (Elastic Container Registry) Lifecycle Policy resources.
---

# List Resource: aws_ecr_lifecycle_policy

Lists ECR (Elastic Container Registry) Lifecycle Policy resources.
Only repositories with an attached lifecycle policy are returned.

## Example Usage

```terraform
list "aws_ecr_lifecycle_policy" "example" {
  provider = aws
}
```

## Argument Reference

This list resource supports the following arguments:

* `account` - (Optional) AWS account ID to query. Defaults to the account of the provider's credentials. See the [Per-Resource Account guide](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/guides/per-resource-account).
* `region` - (Optional) Region to query. Defaults to provider region.
