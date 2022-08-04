---
subcategory: "Neptune"
layout: "aws"
page_title: "AWS: aws_neptune_global_cluster"
description: |-
  Provides an Neptune Global Cluster Resource
---

# Resource: aws_neptune_global_cluster

Manages a Neptune Global Cluster. A global cluster consists of one primary region and up to five read-only secondary regions. You issue write operations directly to the primary cluster in the primary region and Amazon Neptune automatically replicates the data to the secondary regions using dedicated infrastructure.

More information about Neptune Global Clusters can be found in the [Neptune User Guide](https://docs.aws.amazon.com/neptune/latest/userguide/neptune-global-database.html).

## Example Usage

### New Neptune Global Cluster

```terraform
provider "aws" {
  alias  = "primary"
  region = "us-east-2"
}

provider "aws" {
  alias  = "secondary"
  region = "us-east-1"
}

resource "aws_neptune_global_cluster" "example" {
  global_cluster_identifier = "global-test"
  engine                    = "neptune"
  engine_version            = "1.2.0.0"
}

resource "aws_neptune_cluster" "primary" {
  provider                  = aws.primary
  engine                    = aws_neptune_global_cluster.example.engine
  engine_version            = aws_neptune_global_cluster.example.engine_version
  cluster_identifier        = "test-primary-cluster"
  global_cluster_identifier = aws_neptune_global_cluster.example.id
  neptune_subnet_group_name = "default"
}

resource "aws_neptune_cluster_instance" "primary" {
  provider                  = aws.primary
  engine                    = aws_neptune_global_cluster.example.engine
  engine_version            = aws_neptune_global_cluster.example.engine_version
  identifier                = "test-primary-cluster-instance"
  cluster_identifier        = aws_neptune_cluster.primary.id
  instance_class            = "db.r5.large"
  neptune_subnet_group_name = "default"
}

resource "aws_neptune_cluster" "secondary" {
  provider                  = aws.secondary
  engine                    = aws_neptune_global_cluster.example.engine
  engine_version            = aws_neptune_global_cluster.example.engine_version
  cluster_identifier        = "test-secondary-cluster"
  global_cluster_identifier = aws_neptune_global_cluster.example.id
  neptune_subnet_group_name = "default"
}

resource "aws_neptune_cluster_instance" "secondary" {
  provider                  = aws.secondary
  engine                    = aws_neptune_global_cluster.example.engine
  engine_version            = aws_neptune_global_cluster.example.engine_version
  identifier                = "test-secondary-cluster-instance"
  cluster_identifier        = aws_neptune_cluster.secondary.id
  instance_class            = "db.r5.large"
  neptune_subnet_group_name = "default"

  depends_on = [
    aws_neptune_cluster_instance.primary
  ]
}
```

### New Global Cluster From Existing DB Cluster

```terraform
resource "aws_neptune_cluster" "example" {
  # ... other configuration ...

  # NOTE: Using this DB Cluster to create a Global Cluster, the
  # global_cluster_identifier attribute will become populated and
  # Terraform will begin showing it as a difference. Do not configure:
  # global_cluster_identifier = aws_neptune_global_cluster.example.id
  # as it creates a circular reference. Use ignore_changes instead.
  lifecycle {
    ignore_changes = [global_cluster_identifier]
  }
}

resource "aws_neptune_global_cluster" "example" {
  global_cluster_identifier    = "example"
  source_db_cluster_identifier = aws_neptune_cluster.example.arn
}
```

## Argument Reference

The following arguments are supported:

* `global_cluster_identifier` - (Required, Forces new resources) The global cluster identifier.
* `deletion_protection` - (Optional) If the Global Cluster should have deletion protection enabled. The database can't be deleted when this value is set to `true`. The default is `false`.
* `engine` - (Optional, Forces new resources) Name of the database engine to be used for this DB cluster. Terraform will only perform drift detection if a configuration value is provided. Current Valid values: `neptune`. Conflicts with `source_db_cluster_identifier`.
* `engine_version` - (Optional) Engine version of the global database. Upgrading the engine version will result in all cluster members being immediately updated and will.
    * **NOTE:** Upgrading major versions is not supported.
* `source_db_cluster_identifier` - (Optional) Amazon Resource Name (ARN) to use as the primary DB Cluster of the Global Cluster on creation. Terraform cannot perform drift detection of this value.
* `storage_encrypted` - (Optional, Forces new resources) Specifies whether the DB cluster is encrypted. The default is `false` unless `source_db_cluster_identifier` is specified and encrypted. Terraform will only perform drift detection if a configuration value is provided.

### Timeouts

The `timeouts` block allows you to specify [timeouts](https://www.terraform.io/docs/configuration/blocks/resources/syntax.html#operation-timeouts) for certain actions:

* `create` - (Defaults to 5 mins) Used when creating the Global Cluster
* `update` - (Defaults to 5 mins) Used when updating the Global Cluster members (time is per member)
* `delete` - (Defaults to 5 mins) Used when deleting the Global Cluster members (time is per member)

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - Global Cluster Amazon Resource Name (ARN)
* `global_cluster_members` - Set of objects containing Global Cluster members.
    * `db_cluster_arn` - Amazon Resource Name (ARN) of member DB Cluster.
    * `is_writer` - Whether the member is the primary DB Cluster.
* `global_cluster_resource_id` - AWS Region-unique, immutable identifier for the global database cluster. This identifier is found in AWS CloudTrail log entries whenever the AWS KMS key for the DB cluster is accessed.
* `id` - Neptune Global Cluster.

## Import

`aws_neptune_global_cluster` can be imported by using the Global Cluster identifier, e.g.

```
$ terraform import aws_neptune_global_cluster.example example
```

Certain resource arguments, like `source_db_cluster_identifier`, do not have an API method for reading the information after creation. If the argument is set in the Terraform configuration on an imported resource, Terraform will always show a difference. To workaround this behavior, either omit the argument from the Terraform configuration or use [`ignore_changes`](https://www.terraform.io/docs/configuration/meta-arguments/lifecycle.html#ignore_changes) to hide the difference, e.g.

```terraform
resource "aws_neptune_global_cluster" "example" {
  # ... other configuration ...

  # There is no API for reading source_db_cluster_identifier
  lifecycle {
    ignore_changes = [source_db_cluster_identifier]
  }
}
```
