---
subcategory: "RDS (Relational Database)"
layout: "aws"
page_title: "AWS: aws_rds_cluster_instance"
description: |-
  Provides an RDS cluster instance data source.
---

# Data Source: aws_rds_cluster_instance

Provides information about an RDS cluster instance.

## Example Usage

```terraform
data "aws_rds_cluster_instance" "clusterInstanceName" {
  identifier = "clusterInstanceName"
}
```

## Argument Reference

This data source supports the following arguments:

* `identifier` - (Required) The unique identifier of the RDS cluster instance.

## Attribute Reference

See the [RDS Cluster Instance Resource](/docs/providers/aws/r/rds_cluster_instance.html) for details on the
returned attributes - they are identical for all attributes, except the `tags_all`. If you need to get the tags for this resource, use the attribute `tags` as described below.

* `tags` - A map of tags assigned to the resource.
