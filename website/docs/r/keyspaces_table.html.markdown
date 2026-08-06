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

* `keyspace_name` - (Required) The name of the keyspace that the table is going to be created in.
* `table_name` - (Required) The name of the table.

The following arguments are optional:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `auto_scaling_specification` - (Optional) Specifies the auto scaling settings for a table in provisioned capacity mode. Can only be set when `capacity_specification.throughput_mode` is `PROVISIONED`. More information can be found in the [Developer Guide](https://docs.aws.amazon.com/keyspaces/latest/devguide/autoscaling.html).
* `capacity_specification` - (Optional) Specifies the read/write throughput capacity mode for the table.
* `client_side_timestamps` - (Optional) Enables client-side timestamps for the table. By default, the setting is disabled.
* `comment` - (Optional) A description of the table.
* `default_time_to_live` - (Optional) The default Time to Live setting in seconds for the table. More information can be found in the [Developer Guide](https://docs.aws.amazon.com/keyspaces/latest/devguide/TTL-how-it-works.html#ttl-howitworks_default_ttl).
* `encryption_specification` - (Optional) Specifies how the encryption key for encryption at rest is managed for the table. More information can be found in the [Developer Guide](https://docs.aws.amazon.com/keyspaces/latest/devguide/EncryptionAtRest.html).
* `point_in_time_recovery` - (Optional) Specifies if point-in-time recovery is enabled or disabled for the table. More information can be found in the [Developer Guide](https://docs.aws.amazon.com/keyspaces/latest/devguide/PointInTimeRecovery.html).
* `schema_definition` - (Optional) Describes the schema of the table.
* `tags` - (Optional) A map of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.
* `ttl` - (Optional) Enables Time to Live custom settings for the table. More information can be found in the [Developer Guide](https://docs.aws.amazon.com/keyspaces/latest/devguide/TTL.html).

~> **Note:** We recommend using `lifecycle` [`ignore_changes`](https://developer.hashicorp.com/terraform/language/meta-arguments/lifecycle#ignore_changes) for `capacity_specification[0].read_capacity_units` and `capacity_specification[0].write_capacity_units` if `auto_scaling_specification` is set, since Amazon Keyspaces continually adjusts those values in the background. See the [Auto Scaling example](#auto-scaling) above.

~> **Note:** Reading back `auto_scaling_specification` requires the `application-autoscaling:DescribeScalableTargets` and `application-autoscaling:DescribeScalingPolicies` IAM permissions, in addition to the usual `keyspaces:*` permissions, because Amazon Keyspaces auto scaling is implemented on top of Application Auto Scaling. See the [Developer Guide](https://docs.aws.amazon.com/keyspaces/latest/devguide/autoscaling.html) for details. When `auto_scaling_specification` is managed, missing these permissions is a hard error, because the setting cannot be refreshed or checked for drift. Tables that do not manage `auto_scaling_specification` are unaffected.

~> **Note:** Removing the `auto_scaling_specification` block from your configuration disables auto scaling on the table (it does not retain the existing settings). Because of this, a table [imported](#import) with auto scaling enabled but no block in configuration will plan to disable auto scaling until you add the block.

~> **Note:** Changing `capacity_specification.throughput_mode` from `PAY_PER_REQUEST` to `PROVISIONED` requires `read_capacity_units` and `write_capacity_units`. If those units are held under `ignore_changes` (as recommended when auto scaling is enabled), perform the `throughput_mode` change in a separate apply *before* adding `ignore_changes`; otherwise the plan is rejected because the required units are not part of the planned change.

The `auto_scaling_specification` object takes the following arguments:

* `read_capacity_auto_scaling` - (Optional) The auto scaling settings for the table's read capacity.
* `write_capacity_auto_scaling` - (Optional) The auto scaling settings for the table's write capacity.

The `read_capacity_auto_scaling` and `write_capacity_auto_scaling` objects take the following arguments:

* `auto_scaling_disabled` - (Optional) Whether auto scaling is disabled for the table. Defaults to `false`.
* `maximum_units` - (Required when auto scaling is enabled) The maximum level of throughput the table should always be ready to support. Must be between 1 and the max throughput per second quota for the account (40,000 by default).
* `minimum_units` - (Required when auto scaling is enabled) The minimum level of throughput the table should always be ready to support. Must be between 1 and the max throughput per second quota for the account (40,000 by default).
* `target_tracking_scaling_policy_configuration` - (Required when auto scaling is enabled) Describes a target tracking scaling policy configuration, the only scaling policy type Amazon Keyspaces auto scaling currently supports.

The `target_tracking_scaling_policy_configuration` object takes the following arguments:

* `disable_scale_in` - (Optional) A boolean that specifies if scale-in is disabled or enabled for the table. Defaults to `false`, meaning capacity can be automatically scaled down.
* `scale_in_cooldown` - (Optional) The cooldown period in seconds between scale-in activities. Defaults to `0`.
* `scale_out_cooldown` - (Optional) The cooldown period in seconds between scale-out activities. Defaults to `0`.
* `target_value` - (Required when auto scaling is enabled) The target utilization rate of the table, as a percentage between `20` and `90`.

The `capacity_specification` object takes the following arguments:

* `read_capacity_units` - (Optional) The throughput capacity specified for read operations defined in read capacity units (RCUs).
* `throughput_mode` - (Optional) The read/write throughput capacity mode for a table. Valid values: `PAY_PER_REQUEST`, `PROVISIONED`. The default value is `PAY_PER_REQUEST`.
* `write_capacity_units` - (Optional) The throughput capacity specified for write operations defined in write capacity units (WCUs).

The `client_side_timestamps` object takes the following arguments:

* `status` - (Required) Shows how to enable client-side timestamps settings for the specified table. Valid values: `ENABLED`.

The `comment` object takes the following arguments:

* `message` - (Required) A description of the table.

The `encryption_specification` object takes the following arguments:

* `kms_key_identifier` - (Optional) The Amazon Resource Name (ARN) of the customer managed KMS key.
* `type` - (Optional) The encryption option specified for the table. Valid values: `AWS_OWNED_KMS_KEY`, `CUSTOMER_MANAGED_KMS_KEY`. The default value is `AWS_OWNED_KMS_KEY`.

The `point_in_time_recovery` object takes the following arguments:

* `status` - (Optional) Valid values: `ENABLED`, `DISABLED`. The default value is `DISABLED`.

The `schema_definition` object takes the following arguments:

* `column` - (Required) The regular columns of the table.
* `partition_key` - (Required) The columns that are part of the partition key of the table .
* `clustering_key` - (Required) The columns that are part of the clustering key of the table.
* `static_column` - (Required) The columns that have been defined as `STATIC`. Static columns store values that are shared by all rows in the same partition.

The `column` object takes the following arguments:

* `name` - (Required) The name of the column.
* `type` - (Required) The data type of the column. See the [Developer Guide](https://docs.aws.amazon.com/keyspaces/latest/devguide/cql.elements.html#cql.data-types) for a list of available data types.

The `partition_key` object takes the following arguments:

* `name` - (Required) The name of the partition key column.

The `clustering_key` object takes the following arguments:

* `name` - (Required) The name of the clustering key column.
* `order_by` - (Required) The order modifier. Valid values: `ASC`, `DESC`.

The `static_column` object takes the following arguments:

* `name` - (Required) The name of the static column.

The `ttl` object takes the following arguments:

* `status` - (Optional) Valid values: `ENABLED`.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - The ARN of the table.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

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
