---
subcategory: "RDS (Relational Database)"
layout: "aws"
page_title: "AWS: aws_rds_global_clusters"
description: |-
  Terraform data source for listing AWS RDS Global Clusters.
---

# Data Source: aws_rds_global_clusters

Use this data source to discover all RDS Global Clusters in the current AWS account.

## Example Usage

### Basic Usage

```terraform
data "aws_rds_global_clusters" "all" {}
```

### Filter by Engine

```terraform
data "aws_rds_global_clusters" "postgresql" {
  filter {
    name   = "engine"
    values = ["aurora-postgresql"]
  }
}
```

### Chain with Singular Data Source

```terraform
data "aws_rds_global_clusters" "discovered" {
  filter {
    name   = "engine"
    values = ["aurora-postgresql"]
  }
}

data "aws_rds_global_cluster" "detailed" {
  for_each   = toset(data.aws_rds_global_clusters.discovered.global_cluster_identifiers)
  identifier = each.value
}

output "all_global_endpoints" {
  value = { for k, v in data.aws_rds_global_cluster.detailed : k => v.endpoint }
}
```

## Argument Reference

The following arguments are optional:

* `filter` - (Optional) One or more configuration blocks for client-side filtering. Detailed below.

### filter Argument Reference

* `name` - (Required) Name of the field to filter by. Supported values: `engine`, `engine_version`, `status`, `global_cluster_identifier`, `database_name`, `storage_encrypted`, `deletion_protection`.
* `values` - (Required) List of values to match against.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `id` - AWS region.
* `global_cluster_identifiers` - List of user-defined global cluster identifiers.
* `global_cluster_arns` - List of ARNs for the global clusters.
