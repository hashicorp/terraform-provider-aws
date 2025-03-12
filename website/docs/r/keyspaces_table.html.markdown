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

## Argument Reference

The following arguments are required:

* `keyspace_name` - (Required) The name of the keyspace that the table is going to be created in.
* `table_name` - (Required) The name of the table.

The following arguments are optional:

* `auto_scaling_specification` - (Optional) Specifies the autoscaling settings for a table in provisioned capacity mode.
* `capacity_specification` - (Optional) Specifies the read/write throughput capacity mode for the table.
* `client_side_timestamps` - (Optional) Enables client-side timestamps for the table. By default, the setting is disabled.
* `comment` - (Optional) A description of the table.
* `default_time_to_live` - (Optional) The default Time to Live setting in seconds for the table. More information can be found in the [Developer Guide](https://docs.aws.amazon.com/keyspaces/latest/devguide/TTL-how-it-works.html#ttl-howitworks_default_ttl).
* `encryption_specification` - (Optional) Specifies how the encryption key for encryption at rest is managed for the table. More information can be found in the [Developer Guide](https://docs.aws.amazon.com/keyspaces/latest/devguide/EncryptionAtRest.html).
* `point_in_time_recovery` - (Optional) Specifies if point-in-time recovery is enabled or disabled for the table. More information can be found in the [Developer Guide](https://docs.aws.amazon.com/keyspaces/latest/devguide/PointInTimeRecovery.html).
* `schema_definition` - (Optional) Describes the schema of the table.
* `tags` - (Optional) A map of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.
* `ttl` - (Optional) Enables Time to Live custom settings for the table. More information can be found in the [Developer Guide](https://docs.aws.amazon.com/keyspaces/latest/devguide/TTL.html).

The `auto_scaling_specification` object takes the following arguments:

* `read_capacity_auto_scaling` - (Optional) The autoscaling settings for the table's read capacity.
* `write_capacity_auto_scaling` - (Optional) The autoscaling settings for the table's write capacity.

The `read_capacity_auto_scaling` and `write_capacity_auto_scaling` object takes the following arguments:

* `auto_scaling_disabled` - (Optional) Enables autoscaling for the table if set to false.
* `maximum_units` - (Optional) Maximum amount of throughput to provision. The value must be between 1 and the max throughput per second quota for your account. Default is `40000`
* `minimum_units` - (Optional) Minimum amount of throughput to provision. The value must be between 1 and the max throughput per second quota for your account. Default is `40000`
* `scaling_policy` - (Required) The auto scaling target is the provisioned capacity of the table.

The `scaling_policy` object takes the following arguments:

* `target_tracking_scaling_policy_configuration` - (Optional) Target tracking policy.

The `target_tracking_scaling_policy_configuration` object takes the following arguments:

* `target_value` - (Required) Target value for the target tracking autoscaling policy. Must be between 20 and 90.
* `disable_scale_in` - (Optional) Specifies if scale-in is enabled.
* `scale_in_cooldown` - (Optional) Cooldown period in seconds between scaling activities. Default is 0
* `scale_out_cooldown` - (Optional) Cooldown period in seconds between scaling activities. Default is 0

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
