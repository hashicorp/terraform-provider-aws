---
subcategory: "RDS (Relational Database)"
layout: "aws"
page_title: "AWS: aws_rds_shard_group"
description: |-
  Terraform resource for managing an Amazon Aurora Limitless Database DB shard group.
---

# Resource: aws_rds_shard_group

Terraform resource for managing an Amazon Aurora Limitless Database DB shard group

## Example Usage

### Basic Usage

```terraform
resource "aws_rds_cluster" "example" {
  cluster_identifier                    = "example-limitless-cluster"
  engine                                = "aurora-postgresql"
  engine_version                        = "16.6-limitless"
  engine_mode                           = ""
  storage_type                          = "aurora-iopt1"
  cluster_scalability_type              = "limitless"
  master_username                       = "foo"
  master_password                       = "must_be_eight_characters"
  performance_insights_enabled          = true
  performance_insights_retention_period = 31
  enabled_cloudwatch_logs_exports       = ["postgresql"]
  monitoring_interval                   = 5
  monitoring_role_arn                   = aws_iam_role.example.arn
}

resource "aws_rds_shard_group" "example" {
  db_shard_group_identifier = "example-shard-group"
  db_cluster_identifier     = aws_rds_cluster.example.id
  max_acu                   = 1200
}
```

## Argument Reference

For more detailed documentation about each argument, refer to the [AWS official documentation](https://docs.aws.amazon.com/cli/latest/reference/rds/create-integration.html).

This resource supports the following arguments:

* `compute_redundancy` - (Optional) Specifies whether to create standby DB shard groups for the DB shard group. Valid values are:
    * `0` - Creates a DB shard group without a standby DB shard group. This is the default value.
    * `1` - Creates a DB shard group with a standby DB shard group in a different Availability Zone (AZ).
    * `2` - Creates a DB shard group with two standby DB shard groups in two different AZs.
* `db_cluster_identifier` - (Required) The name of the primary DB cluster for the DB shard group.
* `db_shard_group_identifier` - (Required) The name of the DB shard group.
* `max_acu` - (Required) The maximum capacity of the DB shard group in Aurora capacity units (ACUs).
* `min_acu` - (Optional) The minimum capacity of the DB shard group in Aurora capacity units (ACUs).
* `publicly_accessible` - (Optional) Indicates whether the DB shard group is publicly accessible.
* `tags` - (Optional) Key-value map of resource tags. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the shard group.
* `db_shard_group_resource_id` - The AWS Region-unique, immutable identifier for the DB shard group.
* `endpoint` - The connection endpoint for the DB shard group.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `45m`)
* `update` - (Default `45m`)
* `delete` - (Default `45m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import shard group using the `db_shard_group_identifier`. For example:

```terraform
import {
  to = aws_rds_shard_group.example
  id = "example-shard-group"
}
```

Using `terraform import`, import shard group using the `db_shard_group_identifier`. For example:

```console
% terraform import aws_rds_shard_group.example example-shard-group
```
