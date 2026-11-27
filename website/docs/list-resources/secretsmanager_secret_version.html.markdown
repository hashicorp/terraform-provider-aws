---
subcategory: "Secrets Manager"
layout: "aws"
page_title: "AWS: aws_secretsmanager_secret_version"
description: |-
  Lists Secrets Manager secret versions.
---

# List Resource: aws_secretsmanager_secret_version

Lists Secrets Manager secret versions for a secret.

## Example Usage

```terraform
list "aws_secretsmanager_secret_version" "example" {
  provider = aws

  config {
    secret_id = aws_secretsmanager_secret.example.id
  }
}
```

## Argument Reference

This list resource supports the following arguments:

* `account` - (Optional) AWS account ID to query. Defaults to the account of the provider's credentials. See the [Per-Resource Account guide](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/guides/per-resource-account).
* `region` - (Optional) Region to query. Defaults to provider region.
* `secret_id` - (Required) ARN or name of the secret.
