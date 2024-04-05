---
subcategory: "Redshift"
layout: "aws"
page_title: "AWS: aws_redshift_orderable_cluster"
description: |-
  Information about RDS orderable DB instances.
---

# Data Source: aws_redshift_orderable_cluster

Information about Redshift Orderable Clusters and valid parameter combinations.

## Example Usage

```terraform
data "aws_redshift_orderable_cluster" "test" {
  cluster_type         = "multi-node"
  preferred_node_types = ["dc2.large", "ds2.xlarge"]
}
```

## Argument Reference

This data source supports the following arguments:

* `cluster_type` - (Optional) Reshift Cluster typeE.g., `multi-node` or `single-node`
* `cluster_version` - (Optional) Redshift Cluster versionE.g., `1.0`
* `node_type` - (Optional) Redshift Cluster node typeE.g., `dc2.8xlarge`
* `preferred_node_types` - (Optional) Ordered list of preferred Redshift Cluster node types. The first match in this list will be returned. If no preferred matches are found and the original search returned more than one result, an error is returned.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `availability_zones` - List of Availability Zone names where the Redshift Cluster is available.
