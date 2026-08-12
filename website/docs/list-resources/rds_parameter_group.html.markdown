---
subcategory: "RDS (Relational Database)"
layout: "aws"
page_title: "AWS: aws_rds_parameter_group"
description: |-
  Lists RDS (Relational Database) Parameter Group resources.
---

# List Resource: aws_rds_parameter_group

Lists RDS (Relational Database) Parameter Group resources.

## Example Usage

```terraform
list "aws_rds_parameter_group" "example" {
  provider = aws
}
```

## Argument Reference

This list resource supports the following arguments:

* `region` - (Optional) Region to query. Defaults to provider region.
