---
subcategory: "RDS (Relational Database)"
layout: "aws"
page_title: "AWS: aws_rds_clusters"
description: |-
  Terraform data source for managing an AWS RDS (Relational Database) Clusters.
---

# Data Source: aws_rds_clusters

Terraform data source for managing an AWS RDS (Relational Database) Clusters.

## Example Usage

### Basic Usage

```terraform
data "aws_rds_clusters" "example" {
  filter {
    name   = "engine"
    values = ["aurora-postgresql"]
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

* `cluster_arns` - Set of cluster ARNs of the matched RDS clusters.
* `cluster_identifiers` - Set of ARNs of cluster identifiers of the matched RDS clusters.
