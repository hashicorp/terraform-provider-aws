---
subcategory: "Secrets Manager"
layout: "aws"
page_title: "AWS: aws_secretsmanager_secret_policy"
description: |-
  Lists Secrets Manager Secret Policy resources.
---

# List Resource: aws_secretsmanager_secret_policy

Lists Secrets Manager Secret Policy resources.

## Example Usage

```terraform
list "aws_secretsmanager_secret_policy" "example" {
  provider = aws
}
```

## Argument Reference

This list resource supports the following arguments:

* `account` - (Optional) AWS account ID to query. Defaults to the account of the provider's credentials. See the [Per-Resource Account guide](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/guides/per-resource-account).
* `region` - (Optional) Region to query. Defaults to provider region.
