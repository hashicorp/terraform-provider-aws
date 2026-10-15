---
subcategory: "RDS (Relational Database)"
layout: "aws"
page_title: "AWS: aws_db_subnet_group"
description: |-
  Lists RDS DB Subnet Group resources.
---

# List Resource: aws_db_subnet_group

Lists RDS DB Subnet Group resources.

## Example Usage

```terraform
list "aws_db_subnet_group" "example" {
  provider = aws
}
```

## Argument Reference

This list resource supports the following arguments:

* `region` - (Optional) Region to query. Defaults to provider region.
