---
subcategory: "RDS (Relational Database)"
layout: "aws"
page_title: "AWS: aws_rds_cluster_parameter_group"
description: |-
  Information about an RDS cluster parameter group.
---

# Data Source: aws_rds_cluster_parameter_group

Information about an RDS cluster parameter group.

## Example Usage

```terraform
data "aws_rds_cluster_parameter_group" "test" {
  name = "default.postgres15"
}
```

## Argument Reference

The following arguments are required:

* `name` - (Required) DB cluster parameter group name.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `arn` - ARN of the cluster parameter group.
* `family` - Family of the cluster parameter group.
* `description` - Description of the cluster parameter group.
