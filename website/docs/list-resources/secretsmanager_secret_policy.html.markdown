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

* `region` - (Optional) Region to query. Defaults to provider region.
