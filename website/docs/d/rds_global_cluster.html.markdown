---
subcategory: "RDS (Relational Database)"
layout: "aws"
page_title: "AWS: aws_rds_global_cluster"
description: |-
  Terraform data source for managing an AWS RDS (Relational Database) Global Cluster.
---

# Data Source: aws_rds_global_cluster

Terraform data source for managing an AWS RDS (Relational Database) Global Cluster.

## Example Usage

### Basic Usage

```terraform
data "aws_rds_global_cluster" "example" {
  identifier = aws_rds_global_cluster.test.global_cluster_identifier
}
```

## Argument Reference

The following arguments are required:

* `identifier` - (Required) The global cluster identifier of the RDS global cluster.

The following arguments are optional:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `arn` - RDS Global Cluster Amazon Resource Name (ARN)
* `database_name` - Name of the automatically created database on cluster creation.
* `deletion_protection` -  If the Global Cluster should have deletion protection enabled. The database can't be deleted when this value is set to `true`.
* `endpoint` - The endpoint for the Global Cluster.
* `engine` - Name of the database engine.
* `engine_lifecycle_support` - The current lifecycle support status of the database engine for this Global Cluster.
* `engine_version` -   Version of the database engine for this Global Cluster.
* `storage_encrypted` - Whether the DB cluster is encrypted.
* `members` -  Set of objects containing Global Cluster members.
    * `db_cluster_arn` - Amazon Resource Name (ARN) of member DB Cluster
    * `is_writer` - Whether the member is the primary DB Cluster
* `resource_id` - AWS Region-unique, immutable identifier for the global database cluster.
* `tags` - A map of tags to assigned to the Global Cluster.
