---
subcategory: "RDS (Relational Database)"
layout: "aws"
page_title: "AWS: aws_rds_global_cluster"
description: |-
  Terraform data source for retrieving information about an AWS RDS Global Cluster.
---

# Data Source: aws_rds_global_cluster

Terraform data source for retrieving information about an AWS RDS Global Cluster.

## Example Usage

```terraform
data "aws_rds_global_cluster" "example" {
  global_cluster_identifier = "example-global-cluster"
}
```

## Argument Reference

The following arguments are required:

* `global_cluster_identifier` - (Required) Identifier of the RDS Global Cluster.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `arn` - ARN of the RDS Global Cluster.
* `database_name` - Name of the database.
* `deletion_protection` - Whether the Global Cluster has deletion protection enabled.
* `endpoint` - Writer endpoint for the RDS Global Cluster.
* `engine` - Database engine used by the Global Cluster (e.g., `aurora`, `aurora-mysql`, `aurora-postgresql`).
* `engine_lifecycle_support` - Lifecycle support state for the cluster's engine.
* `engine_version` - Engine version of the Aurora global database.
* `engine_version_actual` - Full engine version information, containing version number and additional details.
* `global_cluster_members` - Set of global cluster members (RDS Clusters) that are part of this global cluster. Detailed below.
* `global_cluster_resource_id` - Immutable identifier assigned by AWS for the global cluster.
* `id` - Identifier of the RDS Global Cluster.
* `storage_encrypted` - Whether the Global Cluster is encrypted.
* `tags` - Map of tags assigned to the Global Cluster.

### global_cluster_members

The `global_cluster_members` attribute has these exported attributes:

* `db_cluster_arn` - ARN of the RDS Cluster.
* `is_writer` - Whether the RDS Cluster is the primary/writer cluster.
