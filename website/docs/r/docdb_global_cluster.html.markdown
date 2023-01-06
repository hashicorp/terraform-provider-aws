---
subcategory: "DocDB (DocumentDB)"
layout: "aws"
page_title: "AWS: aws_docdb_global_cluster"
description: |-
  Manages a DocDB Global Cluster
---

# Resource: aws_docdb_global_cluster

Manages an DocumentDB Global Cluster. A global cluster consists of one primary region and up to five read-only secondary regions. You issue write operations directly to the primary cluster in the primary region and Amazon DocumentDB automatically replicates the data to the secondary regions using dedicated infrastructure.

More information about DocumentDB Global Clusters can be found in the [DocumentDB Developer Guide](https://docs.aws.amazon.com/documentdb/latest/developerguide/global-clusters.html).

## Example Usage

### New DocumentDB Global Cluster

```terraform
provider "aws" {
  alias  = "primary"
  region = "us-east-2"
}

provider "aws" {
  alias  = "secondary"
  region = "us-east-1"
}

resource "aws_docdb_global_cluster" "example" {
  global_cluster_identifier = "global-test"
  engine                    = "docdb"
  engine_version            = "4.0.0"
}

resource "aws_docdb_cluster" "primary" {
  provider                  = aws.primary
  engine                    = aws_docdb_global_cluster.example.engine
  engine_version            = aws_docdb_global_cluster.example.engine_version
  cluster_identifier        = "test-primary-cluster"
  master_username           = "username"
  master_password           = "somepass123"
  global_cluster_identifier = aws_docdb_global_cluster.example.id
  db_subnet_group_name      = "default"
}

resource "aws_docdb_cluster_instance" "primary" {
  provider           = aws.primary
  engine             = aws_docdb_global_cluster.example.engine
  identifier         = "test-primary-cluster-instance"
  cluster_identifier = aws_docdb_cluster.primary.id
  instance_class     = "db.r5.large"
}

resource "aws_docdb_cluster" "secondary" {
  provider                  = aws.secondary
  engine                    = aws_docdb_global_cluster.example.engine
  engine_version            = aws_docdb_global_cluster.example.engine_version
  cluster_identifier        = "test-secondary-cluster"
  global_cluster_identifier = aws_docdb_global_cluster.example.id
  db_subnet_group_name      = "default"
}

resource "aws_docdb_cluster_instance" "secondary" {
  provider           = aws.secondary
  engine             = aws_docdb_global_cluster.example.engine
  identifier         = "test-secondary-cluster-instance"
  cluster_identifier = aws_docdb_cluster.secondary.id
  instance_class     = "db.r5.large"

  depends_on = [
    aws_docdb_cluster_instance.primary
  ]
}
```

### New Global Cluster From Existing DB Cluster

```terraform
resource "aws_docdb_cluster" "example" {
  # ... other configuration ...

  # NOTE: Using this DB Cluster to create a Global Cluster, the
  # global_cluster_identifier attribute will become populated and
  # Terraform will begin showing it as a difference. Do not configure:
  # global_cluster_identifier = aws_docdb_global_cluster.example.id
  # as it creates a circular reference. Use ignore_changes instead.
  lifecycle {
    ignore_changes = [global_cluster_identifier]
  }
}

resource "aws_docdb_global_cluster" "example" {
  global_cluster_identifier    = "example"
  source_db_cluster_identifier = aws_docdb_cluster.example.arn
}
```

## Argument Reference

The following arguments are supported:

* `global_cluster_identifier` - (Required, Forces new resources) The global cluster identifier.
* `database_name` - (Optional, Forces new resources) Name for an automatically created database on cluster creation.
* `deletion_protection` - (Optional) If the Global Cluster should have deletion protection enabled. The database can't be deleted when this value is set to `true`. The default is `false`.
* `engine` - (Optional, Forces new resources) Name of the database engine to be used for this DB cluster. Terraform will only perform drift detection if a configuration value is provided. Current Valid values: `docdb`. Defaults to `docdb`. Conflicts with `source_db_cluster_identifier`.
* `engine_version` - (Optional) Engine version of the global database. Upgrading the engine version will result in all cluster members being immediately updated and will.
    * **NOTE:** Upgrading major versions is not supported.
* `source_db_cluster_identifier` - (Optional) Amazon Resource Name (ARN) to use as the primary DB Cluster of the Global Cluster on creation. Terraform cannot perform drift detection of this value.
* `storage_encrypted` - (Optional, Forces new resources) Specifies whether the DB cluster is encrypted. The default is `false` unless `source_db_cluster_identifier` is specified and encrypted. Terraform will only perform drift detection if a configuration value is provided.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - Global Cluster Amazon Resource Name (ARN)
* `global_cluster_members` - Set of objects containing Global Cluster members.
    * `db_cluster_arn` - Amazon Resource Name (ARN) of member DB Cluster.
    * `is_writer` - Whether the member is the primary DB Cluster.
* `global_cluster_resource_id` - AWS Region-unique, immutable identifier for the global database cluster. This identifier is found in AWS CloudTrail log entries whenever the AWS KMS key for the DB cluster is accessed.
* `id` - DocDB Global Cluster.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `5m`)
* `update` - (Default `5m`)
* `delete` - (Default `5m`)

## Import

`aws_docdb_global_cluster` can be imported by using the Global Cluster identifier, e.g.

```
$ terraform import aws_docdb_global_cluster.example example
```

Certain resource arguments, like `source_db_cluster_identifier`, do not have an API method for reading the information after creation. If the argument is set in the Terraform configuration on an imported resource, Terraform will always show a difference. To workaround this behavior, either omit the argument from the Terraform configuration or use [`ignore_changes`](https://www.terraform.io/docs/configuration/meta-arguments/lifecycle.html#ignore_changes) to hide the difference, e.g.

```terraform
resource "aws_docdb_global_cluster" "example" {
  # ... other configuration ...

  # There is no API for reading source_db_cluster_identifier
  lifecycle {
    ignore_changes = [source_db_cluster_identifier]
  }
}
```
