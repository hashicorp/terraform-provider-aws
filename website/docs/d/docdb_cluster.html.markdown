---
subcategory: "DocumentDB"
layout: "aws"
page_title: "AWS: aws_docdb_cluster"
description: |-
  Terraform data source for managing an AWS DocumentDB Cluster.
---

# Data Source: aws_docdb_cluster

Terraform data source for managing an AWS DocumentDB Cluster.

## Example Usage

```terraform
data "aws_docdb_cluster" "example" {
  cluster_identifier = "clusterIdentifier"
}
```

## Argument Reference

The following arguments are required:

* `cluster_identifier` - (Required) Cluster identifier of the DocumentDB cluster.

## Attribute Reference

See the [DocumentDB Cluster Resource](/docs/providers/aws/r/docdb_cluster.html) for details on the
returned attributes - they are identical for all attributes, except the `tags_all`. If you need to get the tags for this resource, use the attribute `tags` as described below.

* `tags` - A map of tags assigned to the resource.
