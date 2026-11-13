---
subcategory: "Keyspaces (for Apache Cassandra)"
layout: "aws"
page_title: "AWS: aws_keyspaces_table"
description: |-
  Provides a Keyspaces Table.
---

# Resource: aws_keyspaces_table

Provides a Keyspaces Table.

More information about Keyspaces tables can be found in the [Keyspaces Developer Guide](https://docs.aws.amazon.com/keyspaces/latest/devguide/working-with-tables.html).

## Example Usage

```terraform
resource "aws_keyspaces_table" "example" {
  keyspace_name = aws_keyspaces_keyspace.example.name
  table_name    = "my_table"

  schema_definition {
    column {
      name = "Message"
      type = "ASCII"
    }

    partition_key {
      name = "Message"
    }
  }
}
```

### Auto Scaling

~> **Note:** We recommend using `lifecycle` [`ignore_changes`](https://www.terraform.io/docs/configuration/meta-arguments/lifecycle.html#ignore_changes) for `capacity_specification` when using `auto_scaling_specification`.

```terraform
resource "aws_keyspaces_table" "example" {
  keyspace_name = aws_keyspaces_keyspace.example.name
  table_name    = "my_table"

  schema_definition {
    column {
      name = "Message"
      type = "ASCII"
    }

    partition_key {
      name = "Message"
    }
  }

  capacity_specification {
    throughput_mode      = "PROVISIONED"
    read_capacity_units  = 5
    write_capacity_units = 5
  }

  auto_scaling_specification {
    read_capacity_auto_scaling {
      minimum_units = 5
      maximum_units = 10
      target_tracking_scaling_policy_configuration {
        target_value = 70
      }
    }
    write_capacity_auto_scaling {
      minimum_units = 5
      maximum_units = 10
      target_tracking_scaling_policy_configuration {
        target_value = 70
      }
    }
  }

  lifecycle {
    ignore_changes = [
      capacity_specification[0].read_capacity_units,
      capacity_specification[0].write_capacity_units,
    ]
  }
}
```

## Argument Reference

The following arguments are required:

* `keyspace_name` - (Required) Name of the keyspace that the table is going to be created in.
* `schema_definition` - (Required) Schema of the table. See [`schema_definition`](#schema_definition-block) below.
* `table_name` - (Required) Name of the table.

The following arguments are optional:

* `auto_scaling_specification` - (Optional) Auto scaling settings for a table in provisioned capacity mode. Can only be set when `capacity_specification.throughput_mode` is `PROVISIONED`. More information can be found in the [Developer Guide](https://docs.aws.amazon.com/keyspaces/latest/devguide/autoscaling.html). See [`auto_scaling_specification`](#auto_scaling_specification-block) below.
* `capacity_specification` - (Optional) Read/write throughput capacity mode for the table. See [`capacity_specification`](#capacity_specification-block) below.
* `client_side_timestamps` - (Optional) Enables client-side timestamps for the table. By default, the setting is disabled. See [`client_side_timestamps`](#client_side_timestamps-block) below.
* `comment` - (Optional) Description of the table. See [`comment`](#comment-block) below.
* `default_time_to_live` - (Optional) Default Time to Live setting in seconds for the table. More information can be found in the [Developer Guide](https://docs.aws.amazon.com/keyspaces/latest/devguide/TTL-how-it-works.html#ttl-howitworks_default_ttl).
* `encryption_specification` - (Optional) Encryption key management for encryption at rest for the table. More information can be found in the [Developer Guide](https://docs.aws.amazon.com/keyspaces/latest/devguide/EncryptionAtRest.html). See [`encryption_specification`](#encryption_specification-block) below.
* `point_in_time_recovery` - (Optional) Enables or disables point-in-time recovery for the table. More information can be found in the [Developer Guide](https://docs.aws.amazon.com/keyspaces/latest/devguide/PointInTimeRecovery.html). See [`point_in_time_recovery`](#point_in_time_recovery-block) below.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `tags` - (Optional) Map of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.
* `ttl` - (Optional) Enables Time to Live custom settings for the table. More information can be found in the [Developer Guide](https://docs.aws.amazon.com/keyspaces/latest/devguide/TTL.html). See [`ttl`](#ttl-block) below.

### `auto_scaling_specification` Block

~> **Note:** We recommend using `lifecycle` [`ignore_changes`](https://developer.hashicorp.com/terraform/language/meta-arguments/lifecycle#ignore_changes) for `capacity_specification[0].read_capacity_units` and `capacity_specification[0].write_capacity_units` if `auto_scaling_specification` is set, since Amazon Keyspaces continually adjusts those values in the background. See the [Auto Scaling example](#auto-scaling) above.

~> **Note:** Reading back `auto_scaling_specification` requires the `application-autoscaling:DescribeScalableTargets` and `application-autoscaling:DescribeScalingPolicies` IAM permissions, in addition to the usual `keyspaces:*` permissions, because Amazon Keyspaces auto scaling is implemented on top of Application Auto Scaling. See the [Developer Guide](https://docs.aws.amazon.com/keyspaces/latest/devguide/autoscaling.html) for details. When `auto_scaling_specification` is managed, missing these permissions is a hard error, because the setting cannot be refreshed or checked for drift. Tables that do not manage `auto_scaling_specification` are unaffected.

~> **Note:** Removing the `auto_scaling_specification` block from your configuration disables auto scaling on the table (it does not retain the existing settings). Because of this, a table [imported](#import) with auto scaling enabled but no block in configuration will plan to disable auto scaling until you add the block.

~> **Note:** Changing `capacity_specification.throughput_mode` from `PAY_PER_REQUEST` to `PROVISIONED` requires `read_capacity_units` and `write_capacity_units`. If those units are held under `ignore_changes` (as recommended when auto scaling is enabled), perform the `throughput_mode` change in a separate apply *before* adding `ignore_changes`; otherwise the plan is rejected because the required units are not part of the planned change.

* `read_capacity_auto_scaling` - (Optional) Auto scaling settings for the table's read capacity. See [`read_capacity_auto_scaling`](#read_capacity_auto_scaling-block) below.
* `write_capacity_auto_scaling` - (Optional) Auto scaling settings for the table's write capacity. See [`write_capacity_auto_scaling`](#write_capacity_auto_scaling-block) below.

#### `read_capacity_auto_scaling` Block

* `auto_scaling_disabled` - (Optional) Whether to disable auto scaling for the table. Defaults to `false`.
* `maximum_units` - (Optional) Maximum level of throughput the table should always be ready to support. Required when auto scaling is enabled. Must be between 1 and the max throughput per second quota for the account (40,000 by default).
* `minimum_units` - (Optional) Minimum level of throughput the table should always be ready to support. Required when auto scaling is enabled. Must be between 1 and the max throughput per second quota for the account (40,000 by default).
* `target_tracking_scaling_policy_configuration` - (Optional) Target tracking scaling policy configuration, the only scaling policy type Amazon Keyspaces auto scaling currently supports. Required when auto scaling is enabled. See [`read_capacity_auto_scaling.target_tracking_scaling_policy_configuration`](#read_capacity_auto_scalingtarget_tracking_scaling_policy_configuration-block) below.

##### `read_capacity_auto_scaling.target_tracking_scaling_policy_configuration` Block

* `disable_scale_in` - (Optional) Whether to disable scale-in for the table. Defaults to `false`, meaning capacity can be automatically scaled down.
* `scale_in_cooldown` - (Optional) Cooldown period in seconds between scale-in activities. Defaults to `0`.
* `scale_out_cooldown` - (Optional) Cooldown period in seconds between scale-out activities. Defaults to `0`.
* `target_value` - (Optional) Target utilization rate of the table, as a percentage between `20` and `90`. Required when auto scaling is enabled.

#### `write_capacity_auto_scaling` Block

* `auto_scaling_disabled` - (Optional) Whether to disable auto scaling for the table. Defaults to `false`.
* `maximum_units` - (Optional) Maximum level of throughput the table should always be ready to support. Required when auto scaling is enabled. Must be between 1 and the max throughput per second quota for the account (40,000 by default).
* `minimum_units` - (Optional) Minimum level of throughput the table should always be ready to support. Required when auto scaling is enabled. Must be between 1 and the max throughput per second quota for the account (40,000 by default).
* `target_tracking_scaling_policy_configuration` - (Optional) Target tracking scaling policy configuration, the only scaling policy type Amazon Keyspaces auto scaling currently supports. Required when auto scaling is enabled. See [`write_capacity_auto_scaling.target_tracking_scaling_policy_configuration`](#write_capacity_auto_scalingtarget_tracking_scaling_policy_configuration-block) below.

##### `write_capacity_auto_scaling.target_tracking_scaling_policy_configuration` Block

* `disable_scale_in` - (Optional) Whether to disable scale-in for the table. Defaults to `false`, meaning capacity can be automatically scaled down.
* `scale_in_cooldown` - (Optional) Cooldown period in seconds between scale-in activities. Defaults to `0`.
* `scale_out_cooldown` - (Optional) Cooldown period in seconds between scale-out activities. Defaults to `0`.
* `target_value` - (Optional) Target utilization rate of the table, as a percentage between `20` and `90`. Required when auto scaling is enabled.

### `capacity_specification` Block

* `read_capacity_units` - (Optional) Throughput capacity specified for read operations defined in read capacity units (RCUs).
* `throughput_mode` - (Optional) Read/write throughput capacity mode for a table. Valid values: `PAY_PER_REQUEST`, `PROVISIONED`. The default value is `PAY_PER_REQUEST`.
* `write_capacity_units` - (Optional) Throughput capacity specified for write operations defined in write capacity units (WCUs).

### `client_side_timestamps` Block

* `status` - (Required) Shows how to enable client-side timestamps settings for the specified table. Valid values: `ENABLED`.

### `comment` Block

* `message` - (Optional) Description of the table.

### `encryption_specification` Block

* `kms_key_identifier` - (Optional) ARN of the customer managed KMS key.
* `type` - (Optional) Encryption option specified for the table. Valid values: `AWS_OWNED_KMS_KEY`, `CUSTOMER_MANAGED_KMS_KEY`. The default value is `AWS_OWNED_KMS_KEY`.

### `point_in_time_recovery` Block

* `status` - (Optional) Valid values: `ENABLED`, `DISABLED`. The default value is `DISABLED`.

### `schema_definition` Block

* `clustering_key` - (Optional) Columns that are part of the clustering key of the table. See [`clustering_key`](#clustering_key-block) below.
* `column` - (Required) Regular columns of the table. See [`column`](#column-block) below.
* `partition_key` - (Required) Columns that are part of the partition key of the table. See [`partition_key`](#partition_key-block) below.
* `static_column` - (Optional) Columns that have been defined as `STATIC`. Static columns store values that are shared by all rows in the same partition. See [`static_column`](#static_column-block) below.

#### `clustering_key` Block

* `name` - (Required) Name of the clustering key column.
* `order_by` - (Required) Order modifier. Valid values: `ASC`, `DESC`.

#### `column` Block

* `name` - (Required) Name of the column.
* `type` - (Required) Data type of the column. See the [Developer Guide](https://docs.aws.amazon.com/keyspaces/latest/devguide/cql.elements.html#cql.data-types) for a list of available data types.

#### `partition_key` Block

* `name` - (Required) Name of the partition key column.

#### `static_column` Block

* `name` - (Required) Name of the static column.

### `ttl` Block

* `status` - (Required) Valid values: `ENABLED`.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the table.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

- `create` - (Default `10m`)
- `update` - (Default `30m`)
- `delete` - (Default `10m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import a table using the `keyspace_name` and `table_name` separated by `/`. For example:

```terraform
import {
  to = aws_keyspaces_table.example
  id = "my_keyspace/my_table"
}
```

Using `terraform import`, import a table using the `keyspace_name` and `table_name` separated by `/`. For example:

```console
% terraform import aws_keyspaces_table.example my_keyspace/my_table
```
