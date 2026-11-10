---
subcategory: "RDS (Relational Database)"
layout: "aws"
page_title: "AWS: aws_db_instance"
description: |-
  Lists RDS (Relational Database) DB Instance resources.
---

# List Resource: aws_db_instance

Lists RDS (Relational Database) DB Instance resources.

## Example Usage

```terraform
list "aws_db_instance" "example" {
  provider = aws
}
```

## Argument Reference

This list resource supports the following arguments:

* `region` - (Optional) Region to query. Defaults to provider region.
