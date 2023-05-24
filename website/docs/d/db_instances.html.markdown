---
subcategory: "RDS (Relational Database)"
layout: "aws"
page_title: "AWS: aws_db_instances"
description: |-
  Terraform data source for listing RDS Database Instances.
---

# Data Source: aws_db_instances

Terraform data source for listing RDS Database Instances.

## Example Usage

### Basic Usage

```terraform
data "aws_db_instances" "example" {
  filter {
    name   = "db-instance-id"
    values = ["my-database-id"]
  }
}
```

## Argument Reference

The following arguments are optional:

* `filter` - (Optional) Configuration block(s) for filtering. Detailed below.

### filter Configuration block

The following arguments are supported by the `filter` configuration block:

* `name` - (Required) Name of the filter field. Valid values can be found in the [RDS DescribeDBClusters API Reference](https://docs.aws.amazon.com/AmazonRDS/latest/APIReference/API_DescribeDBClusters.html).
* `values` - (Required) Set of values that are accepted for the given filter field. Results will be selected if any given value matches.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `instance_arns` - ARNs of the matched RDS instances.
* `instance_identifiers` - Identifiers of the matched RDS instances.
