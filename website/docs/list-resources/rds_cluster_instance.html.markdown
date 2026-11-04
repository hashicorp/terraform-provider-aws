---
subcategory: "RDS (Relational Database)"
layout: "aws"
page_title: "AWS: aws_rds_cluster_instance"
description: |-
  Lists RDS (Relational Database) Cluster Instance resources.
---

# List Resource: aws_rds_cluster_instance

Lists RDS (Relational Database) Cluster Instance resources.

## Example Usage

```terraform
list "aws_rds_cluster_instance" "example" {
  provider = aws
}
```

## Argument Reference

This list resource supports the following arguments:

* `region` - (Optional) Region to query. Defaults to provider region.
