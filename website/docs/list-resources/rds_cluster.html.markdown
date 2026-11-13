---
subcategory: "RDS (Relational Database)"
layout: "aws"
page_title: "AWS: aws_rds_cluster"
description: |-
  Lists RDS (Relational Database) Cluster resources.
---

# List Resource: aws_rds_cluster

Lists RDS (Relational Database) Cluster resources.

## Example Usage

```terraform
list "aws_rds_cluster" "example" {
  provider = aws
}
```

## Argument Reference

This list resource supports the following arguments:

* `region` - (Optional) Region to query. Defaults to provider region.
