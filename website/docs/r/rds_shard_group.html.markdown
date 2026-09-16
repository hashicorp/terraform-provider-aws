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

This resource supports the following arguments:

* `compute_redundancy` - (Optional) Whether to create standby DB shard groups for the DB shard group. Valid values are `0` (no standby DB shard group, the default), `1` (one standby DB shard group in a different Availability Zone), and `2` (two standby DB shard groups in two different Availability Zones).
* `db_cluster_identifier` - (Required) Name of the primary DB cluster for the DB shard group.
* `db_shard_group_identifier` - (Required) Name of the DB shard group.
* `max_acu` - (Required) Maximum capacity of the DB shard group in Aurora capacity units (ACUs).
* `min_acu` - (Optional) Minimum capacity of the DB shard group in Aurora capacity units (ACUs).
* `publicly_accessible` - (Optional) Whether the DB shard group is publicly accessible.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `tags` - (Optional) Key-value map of resource tags. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

For more detailed documentation about each argument, refer to the [AWS official documentation](https://docs.aws.amazon.com/cli/latest/reference/rds/create-shard-group.html).

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the shard group.
* `db_shard_group_resource_id` - AWS Region-unique, immutable identifier for the DB shard group.
* `endpoint` - Connection endpoint for the DB shard group.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

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
