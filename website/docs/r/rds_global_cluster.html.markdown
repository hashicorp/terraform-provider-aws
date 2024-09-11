---
subcategory: "RDS (Relational Database)"
layout: "aws"
page_title: "AWS: aws_rds_global_cluster"
description: |-
  Manages an RDS Global Cluster
---

# Resource: aws_rds_global_cluster

Manages an RDS Global Cluster, which is an Aurora global database spread across multiple regions. The global database contains a single primary cluster with read-write capability, and a read-only secondary cluster that receives data from the primary cluster through high-speed replication performed by the Aurora storage subsystem.

More information about Aurora global databases can be found in the [Aurora User Guide](https://docs.aws.amazon.com/AmazonRDS/latest/AuroraUserGuide/aurora-global-database.html#aurora-global-database-creating).

## Example Usage

### New MySQL Global Cluster

```terraform
resource "aws_rds_global_cluster" "example" {
  global_cluster_identifier = "global-test"
  engine                    = "aurora"
  engine_version            = "5.6.mysql_aurora.1.22.2"
  database_name             = "example_db"
}

resource "aws_rds_cluster" "primary" {
  provider                  = aws.primary
  engine                    = aws_rds_global_cluster.example.engine
  engine_version            = aws_rds_global_cluster.example.engine_version
  cluster_identifier        = "test-primary-cluster"
  master_username           = "username"
  master_password           = "somepass123"
  database_name             = "example_db"
  global_cluster_identifier = aws_rds_global_cluster.example.id
  db_subnet_group_name      = "default"
}

resource "aws_rds_cluster_instance" "primary" {
  provider             = aws.primary
  engine               = aws_rds_global_cluster.example.engine
  engine_version       = aws_rds_global_cluster.example.engine_version
  identifier           = "test-primary-cluster-instance"
  cluster_identifier   = aws_rds_cluster.primary.id
  instance_class       = "db.r4.large"
  db_subnet_group_name = "default"
}

resource "aws_rds_cluster" "secondary" {
  provider                  = aws.secondary
  engine                    = aws_rds_global_cluster.example.engine
  engine_version            = aws_rds_global_cluster.example.engine_version
  cluster_identifier        = "test-secondary-cluster"
  global_cluster_identifier = aws_rds_global_cluster.example.id
  db_subnet_group_name      = "default"

  depends_on = [
    aws_rds_cluster_instance.primary
  ]
}

resource "aws_rds_cluster_instance" "secondary" {
  provider             = aws.secondary
  engine               = aws_rds_global_cluster.example.engine
  engine_version       = aws_rds_global_cluster.example.engine_version
  identifier           = "test-secondary-cluster-instance"
  cluster_identifier   = aws_rds_cluster.secondary.id
  instance_class       = "db.r4.large"
  db_subnet_group_name = "default"
}
```

### New PostgreSQL Global Cluster

```terraform
provider "aws" {
  alias  = "primary"
  region = "us-east-2"
}

provider "aws" {
  alias  = "secondary"
  region = "us-east-1"
}

resource "aws_rds_global_cluster" "example" {
  global_cluster_identifier = "global-test"
  engine                    = "aurora-postgresql"
  engine_version            = "11.9"
  database_name             = "example_db"
}

resource "aws_rds_cluster" "primary" {
  provider                  = aws.primary
  engine                    = aws_rds_global_cluster.example.engine
  engine_version            = aws_rds_global_cluster.example.engine_version
  cluster_identifier        = "test-primary-cluster"
  master_username           = "username"
  master_password           = "somepass123"
  database_name             = "example_db"
  global_cluster_identifier = aws_rds_global_cluster.example.id
  db_subnet_group_name      = "default"
}

resource "aws_rds_cluster_instance" "primary" {
  provider             = aws.primary
  engine               = aws_rds_global_cluster.example.engine
  engine_version       = aws_rds_global_cluster.example.engine_version
  identifier           = "test-primary-cluster-instance"
  cluster_identifier   = aws_rds_cluster.primary.id
  instance_class       = "db.r4.large"
  db_subnet_group_name = "default"
}

resource "aws_rds_cluster" "secondary" {
  provider                  = aws.secondary
  engine                    = aws_rds_global_cluster.example.engine
  engine_version            = aws_rds_global_cluster.example.engine_version
  cluster_identifier        = "test-secondary-cluster"
  global_cluster_identifier = aws_rds_global_cluster.example.id
  skip_final_snapshot       = true
  db_subnet_group_name      = "default"

  depends_on = [
    aws_rds_cluster_instance.primary
  ]
}

resource "aws_rds_cluster_instance" "secondary" {
  provider             = aws.secondary
  engine               = aws_rds_global_cluster.example.engine
  engine_version       = aws_rds_global_cluster.example.engine_version
  identifier           = "test-secondary-cluster-instance"
  cluster_identifier   = aws_rds_cluster.secondary.id
  instance_class       = "db.r4.large"
  db_subnet_group_name = "default"
}
```

### New Global Cluster From Existing DB Cluster

```terraform
resource "aws_rds_cluster" "example" {
  # ... other configuration ...

  # NOTE: Using this DB Cluster to create a Global Cluster, the
  # global_cluster_identifier attribute will become populated and
  # Terraform will begin showing it as a difference. Do not configure:
  # global_cluster_identifier = aws_rds_global_cluster.example.id
  # as it creates a circular reference. Use ignore_changes instead.
  lifecycle {
    ignore_changes = [global_cluster_identifier]
  }
}

resource "aws_rds_global_cluster" "example" {
  force_destroy                = true
  global_cluster_identifier    = "example"
  source_db_cluster_identifier = aws_rds_cluster.example.arn
}
```

### Upgrading Engine Versions

When you upgrade the version of an `aws_rds_global_cluster`, Terraform will attempt to in-place upgrade the engine versions of all associated clusters. Since the `aws_rds_cluster` resource is being updated through the `aws_rds_global_cluster`, you are likely to get an error (`Provider produced inconsistent final plan`). To avoid this, use the `lifecycle` `ignore_changes` meta argument as shown below on the `aws_rds_cluster`.

```terraform
resource "aws_rds_global_cluster" "example" {
  global_cluster_identifier = "kyivkharkiv"
  engine                    = "aurora-mysql"
  engine_version            = "5.7.mysql_aurora.2.07.5"
}

resource "aws_rds_cluster" "primary" {
  allow_major_version_upgrade = true
  apply_immediately           = true
  cluster_identifier          = "odessadnipro"
  database_name               = "totoro"
  engine                      = aws_rds_global_cluster.example.engine
  engine_version              = aws_rds_global_cluster.example.engine_version
  global_cluster_identifier   = aws_rds_global_cluster.example.id
  master_password             = "satsukimae"
  master_username             = "maesatsuki"
  skip_final_snapshot         = true

  lifecycle {
    ignore_changes = [engine_version]
  }
}

resource "aws_rds_cluster_instance" "primary" {
  apply_immediately  = true
  cluster_identifier = aws_rds_cluster.primary.id
  engine             = aws_rds_cluster.primary.engine
  engine_version     = aws_rds_cluster.primary.engine_version
  identifier         = "donetsklviv"
  instance_class     = "db.r4.large"
}
```

## Argument Reference

This resource supports the following arguments:

* `global_cluster_identifier` - (Required, Forces new resources) Global cluster identifier.
* `database_name` - (Optional, Forces new resources) Name for an automatically created database on cluster creation.
* `deletion_protection` - (Optional) If the Global Cluster should have deletion protection enabled. The database can't be deleted when this value is set to `true`. The default is `false`.
* `engine` - (Optional, Forces new resources) Name of the database engine to be used for this DB cluster. Terraform will only perform drift detection if a configuration value is provided. Valid values: `aurora`, `aurora-mysql`, `aurora-postgresql`. Defaults to `aurora`. Conflicts with `source_db_cluster_identifier`.
* `engine_lifecycle_support` - (Optional) The life cycle type for this DB instance. This setting applies only to Aurora PostgreSQL-based global databases. Valid values are `open-source-rds-extended-support`, `open-source-rds-extended-support-disabled`. Default value is `open-source-rds-extended-support`. [Using Amazon RDS Extended Support]: https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/extended-support.html
* `engine_version` - (Optional) Engine version of the Aurora global database. The `engine`, `engine_version`, and `instance_class` (on the `aws_rds_cluster_instance`) must together support global databases. See [Using Amazon Aurora global databases](https://docs.aws.amazon.com/AmazonRDS/latest/AuroraUserGuide/aurora-global-database.html) for more information. By upgrading the engine version, Terraform will upgrade cluster members. **NOTE:** To avoid an `inconsistent final plan` error while upgrading, use the `lifecycle` `ignore_changes` for `engine_version` meta argument on the associated `aws_rds_cluster` resource as shown above in [Upgrading Engine Versions](#upgrading-engine-versions) example.
* `force_destroy` - (Optional) Enable to remove DB Cluster members from Global Cluster on destroy. Required with `source_db_cluster_identifier`.
* `source_db_cluster_identifier` - (Optional) Amazon Resource Name (ARN) to use as the primary DB Cluster of the Global Cluster on creation. Terraform cannot perform drift detection of this value.
* `storage_encrypted` - (Optional, Forces new resources) Specifies whether the DB cluster is encrypted. The default is `false` unless `source_db_cluster_identifier` is specified and encrypted. Terraform will only perform drift detection if a configuration value is provided.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - RDS Global Cluster Amazon Resource Name (ARN)
* `global_cluster_members` - Set of objects containing Global Cluster members.
    * `db_cluster_arn` - Amazon Resource Name (ARN) of member DB Cluster
    * `is_writer` - Whether the member is the primary DB Cluster
* `global_cluster_resource_id` - AWS Region-unique, immutable identifier for the global database cluster. This identifier is found in AWS CloudTrail log entries whenever the AWS KMS key for the DB cluster is accessed
* `id` - RDS Global Cluster identifier

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

- `create` - (Default `30m`)
- `update` - (Default `90m`)
- `delete` - (Default `30m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import `aws_rds_global_cluster` using the RDS Global Cluster identifier. For example:

```terraform
import {
  to = aws_rds_global_cluster.example
  id = "example"
}
```

Using `terraform import`, import `aws_rds_global_cluster` using the RDS Global Cluster identifier. For example:

```console
% terraform import aws_rds_global_cluster.example example
```

Certain resource arguments, like `force_destroy`, only exist within Terraform. If the argument is set in the Terraform configuration on an imported resource, Terraform will show a difference on the first plan after import to update the state value. This change is safe to apply immediately so the state matches the desired configuration.

Certain resource arguments, like `source_db_cluster_identifier`, do not have an API method for reading the information after creation. If the argument is set in the Terraform configuration on an imported resource, Terraform will always show a difference. To workaround this behavior, either omit the argument from the Terraform configuration or use [`ignore_changes`](https://www.terraform.io/docs/configuration/meta-arguments/lifecycle.html#ignore_changes) to hide the difference. For example:

```terraform
resource "aws_rds_global_cluster" "example" {
  # ... other configuration ...

  # There is no API for reading source_db_cluster_identifier
  lifecycle {
    ignore_changes = [source_db_cluster_identifier]
  }
}
```
