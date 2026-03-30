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

* `region` - (Optional) Region to query. Defaults to provider region.
* `secret_id` - (Required) ARN or name of the secret.
