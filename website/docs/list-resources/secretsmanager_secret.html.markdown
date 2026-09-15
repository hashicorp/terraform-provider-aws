---
subcategory: "Secrets Manager"
layout: "aws"
page_title: "AWS: aws_secretsmanager_secret"
description: |-
  Lists Secrets Manager Secret resources.
---

# List Resource: aws_secretsmanager_secret

Lists Secrets Manager Secret resources.

## Example Usage

```terraform
list "aws_secretsmanager_secret" "example" {
  provider = aws
}
```

## Argument Reference

This list resource supports the following arguments:

* `region` - (Optional) Region to query. Defaults to provider region.
