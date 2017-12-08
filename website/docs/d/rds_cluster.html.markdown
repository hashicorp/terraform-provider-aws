---
layout: "aws"
page_title: "AWS: aws_rds_cluster"
sidebar_current: "docs-aws-datasource-rds-cluster"
description: |-
  Provides a RDS cluster data source.
---

# Data Source: aws_rds_cluster

Provides information about a RDS cluster.

## Example Usage

```hcl
data "aws_rds_cluster" "clusterName" {
  name = "clusterName"
}
```

## Argument Reference

The following arguments are supported:

* `cluster_identifier` - (Required) The cluster identifier of the RDS cluster.

## Attributes Reference

See the [RDS Cluster Resource](/docs/providers/aws/r/rds_cluster.html) for details on the
returned attributes - they are identical.
