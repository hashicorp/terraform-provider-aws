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

### Using tags

```terraform
data "aws_db_instances" "example" {
  tags = {
    Env = "test"
  }
}
```

## Argument Reference

The following arguments are optional:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `filter` - (Optional) Configuration block(s) used to filter instances with AWS supported attributes, such as `engine`, `db-cluster-id` or `db-instance-id` for example. Detailed below.
* `tags` - (Optional) Map of tags, each pair of which must exactly match a pair on the desired instances.

### filter Configuration block

The `filter` configuration block supports the following arguments:

* `name` - (Required) Name of the filter field. Valid values can be found in the [RDS DescribeDBClusters API Reference](https://docs.aws.amazon.com/AmazonRDS/latest/APIReference/API_DescribeDBClusters.html) or [RDS DescribeDBInstances API Reference](https://docs.aws.amazon.com/AmazonRDS/latest/APIReference/API_DescribeDBInstances.html).
* `values` - (Required) Set of values that are accepted for the given filter field. Results will be selected if any given value matches.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `instance_arns` - ARNs of the matched RDS instances.
* `instance_identifiers` - Identifiers of the matched RDS instances.
