---
subcategory: "DynamoDB"
layout: "aws"
page_title: "AWS: aws_dynamodb_table"
description: |-
  Provides a DynamoDB table resource
---

# Resource: aws_dynamodb_table

Provides a DynamoDB table resource.

~> **Note:** We recommend using `lifecycle` [`ignore_changes`](https://www.terraform.io/docs/configuration/meta-arguments/lifecycle.html#ignore_changes) for `read_capacity` and/or `write_capacity` if there's [autoscaling policy](/docs/providers/aws/r/appautoscaling_policy.html) attached to the table.

~> **Note:** When using [aws_dynamodb_table_replica](/docs/providers/aws/r/dynamodb_table_replica.html) with this resource, use `lifecycle` [`ignore_changes`](https://www.terraform.io/docs/configuration/meta-arguments/lifecycle.html#ignore_changes) for `replica`, _e.g._, `lifecycle { ignore_changes = [replica] }`.

~> **Note:** If autoscaling creates drift for your `global_secondary_index` blocks and/or more granular `lifecycle` management for GSIs, we recommend using the new **experimental** resource [`aws_dynamodb_global_secondary_index`](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/dynamodb_global_secondary_index).

## DynamoDB Table attributes

Only define attributes on the table object that are going to be used as:

* Table hash key or range key
* LSI or GSI hash key or range key

The DynamoDB API expects attribute structure (name and type) to be passed along when creating or updating GSI/LSIs or creating the initial table. In these cases it expects the Hash / Range keys to be provided. Because these get re-used in numerous places (i.e the table's range key could be a part of one or more GSIs), they are stored on the table object to prevent duplication and increase consistency. If you add attributes here that are not used in these scenarios it can cause an infinite loop in planning.

~> **Note:** When using the [`aws_dynamodb_global_secondary_index`](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/dynamodb_global_secondary_index) resource, you do not need to define the attributes for externally managed GSIs in the `aws_dynamodb_table` resource.

## Example Usage

### Basic Example

The following dynamodb table description models the table and GSI shown in the [AWS SDK example documentation](https://docs.aws.amazon.com/amazondynamodb/latest/developerguide/GSI.html)

```terraform
resource "aws_dynamodb_table" "basic-dynamodb-table" {
  name           = "GameScores"
  billing_mode   = "PROVISIONED"
  read_capacity  = 20
  write_capacity = 20
  hash_key       = "UserId"
  range_key      = "GameTitle"

  attribute {
    name = "UserId"
    type = "S"
  }

  attribute {
    name = "GameTitle"
    type = "S"
  }

  attribute {
    name = "TopScore"
    type = "N"
  }

  ttl {
    attribute_name = "TimeToExist"
    enabled        = true
  }

  global_secondary_index {
    name               = "GameTitleIndex"
    hash_key           = "GameTitle"
    range_key          = "TopScore"
    write_capacity     = 10
    read_capacity      = 10
    projection_type    = "INCLUDE"
    non_key_attributes = ["UserId"]
  }

  tags = {
    Name        = "dynamodb-table-1"
    Environment = "production"
  }
}
```

### Basic Example containing Global Secondary Indexes using Multi-attribute keys pattern

The following dynamodb table description models the table and GSIs shown in the [AWS SDK example documentation](https://docs.aws.amazon.com/amazondynamodb/latest/developerguide/GSI.DesignPattern.MultiAttributeKeys.html)

~> **Note:** Multi-attribute keys for GSIs use the `key_schema` block instead of `hash_key`/`range_key`. The `hash_key` and `range_key` arguments are deprecated in favor of `key_schema`.

```terraform
resource "aws_dynamodb_table" "basic-dynamodb-table" {
  name           = "TournamentMatches"
  billing_mode   = "PROVISIONED"
  read_capacity  = 20
  write_capacity = 20
  hash_key       = "matchId"

  attribute {
    name = "matchId"
    type = "S"
  }

  attribute {
    name = "tournamentId"
    type = "S"
  }

  attribute {
    name = "region"
    type = "S"
  }

  attribute {
    name = "round"
    type = "S"
  }

  attribute {
    name = "bracket"
    type = "S"
  }

  attribute {
    name = "playerId"
    type = "N"
  }

  attribute {
    name = "matchDate"
    type = "S"
  }

  ttl {
    attribute_name = "TimeToExist"
    enabled        = true
  }

  # GSI with multiple HASH keys and multiple RANGE keys using key_schema
  global_secondary_index {
    name = "TournamentRegionIndex"
    key_schema {
      attribute_name = "tournamentId"
      key_type       = "HASH"
    }
    key_schema {
      attribute_name = "region"
      key_type       = "HASH"
    }
    key_schema {
      attribute_name = "round"
      key_type       = "RANGE"
    }
    key_schema {
      attribute_name = "bracket"
      key_type       = "RANGE"
    }
    key_schema {
      attribute_name = "matchId"
      key_type       = "RANGE"
    }
    write_capacity  = 10
    read_capacity   = 10
    projection_type = "ALL"
  }

  # GSI with single HASH key and multiple RANGE keys using key_schema
  global_secondary_index {
    name = "PlayerMatchHistoryIndex"
    key_schema {
      attribute_name = "playerId"
      key_type       = "HASH"
    }
    key_schema {
      attribute_name = "matchDate"
      key_type       = "RANGE"
    }
    key_schema {
      attribute_name = "round"
      key_type       = "RANGE"
    }
    write_capacity  = 10
    read_capacity   = 10
    projection_type = "ALL"
  }

  tags = {
    Name        = "dynamodb-table-1"
    Environment = "production"
  }
}
```

### Global Tables

This resource implements support for [DynamoDB Global Tables V2 (version 2019.11.21)](https://docs.aws.amazon.com/amazondynamodb/latest/developerguide/globaltables.V2.html) via `replica` configuration blocks. For working with [DynamoDB Global Tables V1 (version 2017.11.29)](https://docs.aws.amazon.com/amazondynamodb/latest/developerguide/globaltables.V1.html), see the [`aws_dynamodb_global_table` resource](/docs/providers/aws/r/dynamodb_global_table.html).

~> **Note:** [aws_dynamodb_table_replica](/docs/providers/aws/r/dynamodb_table_replica.html) is an alternate way of configuring Global Tables. Do not use `replica` configuration blocks of `aws_dynamodb_table` together with [aws_dynamodb_table_replica](/docs/providers/aws/r/dynamodb_table_replica.html).

```terraform
resource "aws_dynamodb_table" "example" {
  name             = "example"
  hash_key         = "TestTableHashKey"
  billing_mode     = "PAY_PER_REQUEST"
  stream_enabled   = true
  stream_view_type = "NEW_AND_OLD_IMAGES"

  attribute {
    name = "TestTableHashKey"
    type = "S"
  }

  replica {
    region_name = "us-east-2"
  }

  replica {
    region_name = "us-west-2"
  }
}
```

#### Global Tables with Multi-Region Strong Consistency

A global table configured for Multi-Region strong consistency (MRSC) provides the ability to perform a strongly consistent read with multi-Region scope. Performing a strongly consistent read on an MRSC table ensures you're always reading the latest version of an item, irrespective of the Region in which you're performing the read.

You can configure a MRSC global table with three replicas, or with two replicas and one witness. A witness is a component of a MRSC global table that contains data written to global table replicas, and provides an optional alternative to a full replica while supporting MRSC's availability architecture. You cannot perform read or write operations on a witness. A witness is located in a different Region than the two replicas.

**Note** Please see detailed information, restrictions, caveats etc on the [AWS Support Page](https://docs.aws.amazon.com/amazondynamodb/latest/developerguide/multi-region-strong-consistency-gt.html).

Consistency Mode (`consistency_mode`) on the embedded `replica` allows you to configure consistency mode for Global Tables.

##### Consistency mode with 3 Replicas

```terraform
resource "aws_dynamodb_table" "example" {
  name             = "example"
  hash_key         = "TestTableHashKey"
  billing_mode     = "PAY_PER_REQUEST"
  stream_enabled   = true
  stream_view_type = "NEW_AND_OLD_IMAGES"

  attribute {
    name = "TestTableHashKey"
    type = "S"
  }

  replica {
    region_name      = "us-east-2"
    consistency_mode = "STRONG"
  }

  replica {
    region_name      = "us-west-2"
    consistency_mode = "STRONG"
  }
}
```

##### Consistency Mode with 2 Replicas and Witness Region

```terraform
resource "aws_dynamodb_table" "example" {
  name             = "example"
  hash_key         = "TestTableHashKey"
  billing_mode     = "PAY_PER_REQUEST"
  stream_enabled   = true
  stream_view_type = "NEW_AND_OLD_IMAGES"

  attribute {
    name = "TestTableHashKey"
    type = "S"
  }

  replica {
    region_name      = "us-east-2"
    consistency_mode = "STRONG"
  }

  global_table_witness {
    region_name = "us-west-2"
  }
}
```

### Replica Tagging

You can manage global table replicas' tags in various ways. This example shows using `replica.*.propagate_tags` for the first replica and the `aws_dynamodb_tag` resource for the other.

```terraform
provider "aws" {
  region = "us-west-2"
}

provider "awsalternate" {
  region = "us-east-1"
}

provider "awsthird" {
  region = "us-east-2"
}

data "aws_region" "current" {}

data "aws_region" "alternate" {
  provider = "awsalternate"
}

data "aws_region" "third" {
  provider = "awsthird"
}

resource "aws_dynamodb_table" "example" {
  billing_mode     = "PAY_PER_REQUEST"
  hash_key         = "TestTableHashKey"
  name             = "example-13281"
  stream_enabled   = true
  stream_view_type = "NEW_AND_OLD_IMAGES"

  attribute {
    name = "TestTableHashKey"
    type = "S"
  }

  replica {
    region_name = data.aws_region.alternate.name
  }

  replica {
    region_name    = data.aws_region.third.name
    propagate_tags = true
  }

  tags = {
    Architect = "Eleanor"
    Zone      = "SW"
  }
}

resource "aws_dynamodb_tag" "example" {
  resource_arn = replace(aws_dynamodb_table.example.arn, data.aws_region.current.region, data.aws_region.alternate.name)
  key          = "Architect"
  value        = "Gigi"
}
```

## Argument Reference

The following arguments are required:

* `attribute` - (Required) Set of nested attribute definitions. Only required for `hash_key` and `range_key` attributes. See below.
* `hash_key` - (Required, Forces new resource) Attribute to use as the hash (partition) key. Must also be defined as an `attribute`. See below.
* `name` - (Required) Unique within a region name of the table.

The following arguments are optional:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `billing_mode` - (Optional) Controls how you are charged for read and write throughput and how you manage capacity. The valid values are `PROVISIONED` and `PAY_PER_REQUEST`. Defaults to `PROVISIONED`.
* `deletion_protection_enabled` - (Optional) Enables deletion protection for table. Defaults to `false`.
* `import_table` - (Optional) Import Amazon S3 data into a new table. See below.
* `global_secondary_index` - (Optional) Describe a GSI for the table; subject to the normal limits on the number of GSIs, projected attributes, etc. See below.
* `global_table_witness` - (Optional) Witness Region in a Multi-Region Strong Consistency deployment. **Note** This must be used alongside a single `replica` with `consistency_mode` set to `STRONG`. Other combinations will fail to provision. See below.
* `local_secondary_index` - (Optional, Forces new resource) Describe an LSI on the table; these can only be allocated _at creation_ so you cannot change this definition after you have created the resource. See below.
* `on_demand_throughput` - (Optional) Sets the maximum number of read and write units for the specified on-demand table. See below.
* `point_in_time_recovery` - (Optional) Enable point-in-time recovery options. See below.
* `range_key` - (Optional, Forces new resource) Attribute to use as the range (sort) key. Must also be defined as an `attribute`, see below.
* `read_capacity` - (Optional) Number of read units for this table. If the `billing_mode` is `PROVISIONED`, this field is required.
* `replica` - (Optional) Configuration block(s) with [DynamoDB Global Tables V2 (version 2019.11.21)](https://docs.aws.amazon.com/amazondynamodb/latest/developerguide/globaltables.V2.html) replication configurations. See below.
* `restore_date_time` - (Optional) Time of the point-in-time recovery point to restore.
* `restore_source_name` - (Optional) Name of the table to restore. Must match the name of an existing table.
* `restore_source_table_arn` - (Optional) ARN of the source table to restore. Must be supplied for cross-region restores.
* `restore_to_latest_time` - (Optional) If set, restores table to the most recent point-in-time recovery point.
* `server_side_encryption` - (Optional) Encryption at rest options. AWS DynamoDB tables are automatically encrypted at rest with an AWS-owned Customer Master Key if this argument isn't specified. Must be supplied for cross-region restores. See below.
* `stream_enabled` - (Optional) Whether Streams are enabled.
* `stream_view_type` - (Optional) When an item in the table is modified, StreamViewType determines what information is written to the table's stream.
  Valid values are `KEYS_ONLY`, `NEW_IMAGE`, `OLD_IMAGE`, `NEW_AND_OLD_IMAGES`.
  Only valid when `stream_enabled` is true.
* `table_class` - (Optional) Storage class of the table.
  Valid values are `STANDARD` and `STANDARD_INFREQUENT_ACCESS`.
  Default value is `STANDARD`.
* `tags` - (Optional) A map of tags to populate on the created table. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.
* `ttl` - (Optional) Configuration block for TTL. See below.
* `warm_throughput` - (Optional) Sets the number of warm read and write units for the specified table. See below.
* `write_capacity` - (Optional) Number of write units for this table. If the `billing_mode` is `PROVISIONED`, this field is required.

### `attribute`

* `name` - (Required) Name of the attribute
* `type` - (Required) Attribute type. Valid values are `S` (string), `N` (number), `B` (binary).

### `import_table`

* `input_compression_type` - (Optional) Type of compression to be used on the input coming from the imported table.
  Valid values are `GZIP`, `ZSTD` and `NONE`.
* `input_format` - (Required) The format of the source data.
  Valid values are `CSV`, `DYNAMODB_JSON`, and `ION`.
* `input_format_options` - (Optional) Describe the format options for the data that was imported into the target table.
  There is one value, `csv`.
  See below.
* `s3_bucket_source` - (Required) Values for the S3 bucket the source file is imported from.
  See below.

#### `input_format_options`

* `csv` - (Optional) This block contains the processing options for the CSV file being imported:
    * `delimiter` - (Optional) The delimiter used for separating items in the CSV file being imported.
    * `header_list` - (Optional) List of the headers used to specify a common header for all source CSV files being imported.

#### `s3_bucket_source`

* `bucket` - (Required) The S3 bucket that is being imported from.
* `bucket_owner`- (Optional) The account number of the S3 bucket that is being imported from.
* `key_prefix` - (Optional) The key prefix shared by all S3 Objects that are being imported.

### `global_secondary_index`

* `hash_key` - (Optional, **Deprecated**) Name of the hash key in the index; must be defined as an attribute in the resource. Mutually exclusive with `key_schema`. Use `key_schema` instead.
* `key_schema` - (Optional) Configuration block(s) for the key schema. Mutually exclusive with `hash_key` and `range_key`. Required if `hash_key` is not specified. Supports multi-attribute keys for the [Multi-Attribute Keys design pattern](https://docs.aws.amazon.com/amazondynamodb/latest/developerguide/GSI.DesignPattern.MultiAttributeKeys.html). See below.
* `name` - (Required) Name of the index.
* `non_key_attributes` - (Optional) Only required with `INCLUDE` as a projection type; a list of attributes to project into the index. These do not need to be defined as attributes on the table.
* `on_demand_throughput` - (Optional) Sets the maximum number of read and write units for the specified on-demand index. See below.
* `projection_type` - (Required) One of `ALL`, `INCLUDE` or `KEYS_ONLY` where `ALL` projects every attribute into the index, `KEYS_ONLY` projects into the index only the table and index hash_key and sort_key attributes, `INCLUDE` projects into the index all of the attributes that are defined in `non_key_attributes` in addition to the attributes that `KEYS_ONLY` project.
* `range_key` - (Optional, **Deprecated**) Name of the range key; must be defined as an attribute in the resource. Mutually exclusive with `key_schema`. Use `key_schema` instead.
* `read_capacity` - (Optional) Number of read units for this index. Must be set if billing_mode is set to PROVISIONED.
* `warm_throughput` - (Optional) Sets the number of warm read and write units for this index. See below.
* `write_capacity` - (Optional) Number of write units for this index. Must be set if billing_mode is set to PROVISIONED.

#### `key_schema`

* `attribute_name` - (Required) Name of the attribute; must be defined as an attribute in the resource.
* `key_type` - (Required) The type of key. Valid values are `HASH` (partition key) or `RANGE` (sort key). You can specify up to 4 attributes with `key_type = "HASH"` and up to 4 attributes with `key_type = "RANGE"`.

### `global_table_witness`

* `region_name` - (Required) Name of the AWS Region that serves as a witness for the MRSC global table.

### `local_secondary_index`

* `name` - (Required) Name of the index
* `non_key_attributes` - (Optional) Only required with `INCLUDE` as a projection type; a list of attributes to project into the index. These do not need to be defined as attributes on the table.
* `projection_type` - (Required) One of `ALL`, `INCLUDE` or `KEYS_ONLY` where `ALL` projects every attribute into the index, `KEYS_ONLY` projects  into the index only the table and index hash_key and sort_key attributes ,  `INCLUDE` projects into the index all of the attributes that are defined in `non_key_attributes` in addition to the attributes that that`KEYS_ONLY` project.
* `range_key` - (Required) Name of the range key.

### `on_demand_throughput`

* `max_read_request_units` - (Optional) Maximum number of read request units for the specified table. To specify set the value greater than or equal to 1. To remove set the value to -1.
* `max_write_request_units` - (Optional) Maximum number of write request units for the specified table. To specify set the value greater than or equal to 1. To remove set the value to -1.

### `point_in_time_recovery`

* `enabled` - (Required) Whether to enable point-in-time recovery. It can take 10 minutes to enable for new tables. If the `point_in_time_recovery` block is not provided, this defaults to `false`.
* `recovery_period_in_days` - (Optional) Number of preceding days for which continuous backups are taken and maintained. Default is 35.

### `replica`

* `kms_key_arn` - (Optional) ARN of the CMK that should be used for the AWS KMS encryption.
  This argument should only be used if the key is different from the default KMS-managed DynamoDB key, `alias/aws/dynamodb`.
  **Note:** This attribute will _not_ be populated with the ARN of _default_ keys.
  **Note:** Changing this value will recreate the replica.
* `point_in_time_recovery` - (Optional) Whether to enable Point In Time Recovery for the replica. Default is `false`.
* `deletion_protection_enabled` - (Optional) Whether deletion protection is enabled (true) or disabled (false) on the replica. Default is `false`.
* `propagate_tags` - (Optional) Whether to propagate the global table's tags to a replica.
  Default is `false`.
  Changes to tags only move in one direction: from global (source) to replica.
  Tag drift on a replica will not trigger an update.
  Tag changes on the global table are propagated to replicas.
  Changing from `true` to `false` on a subsequent `apply` leaves replica tags as-is and no longer manages them.
* `region_name` - (Required) Region name of the replica.
* `consistency_mode` - (Optional) Whether this global table will be using `STRONG` consistency mode or `EVENTUAL` consistency mode. Default value is `EVENTUAL`.

### `server_side_encryption`

* `enabled` - (Required) Whether or not to enable encryption at rest using an AWS managed KMS customer master key (CMK). If `enabled` is `false` then server-side encryption is set to AWS-_owned_ key (shown as `DEFAULT` in the AWS console). Potentially confusingly, if `enabled` is `true` and no `kms_key_arn` is specified then server-side encryption is set to the _default_ KMS-_managed_ key (shown as `KMS` in the AWS console). The [AWS KMS documentation](https://docs.aws.amazon.com/kms/latest/developerguide/concepts.html) explains the difference between AWS-_owned_ and KMS-_managed_ keys.
* `kms_key_arn` - (Optional) ARN of the CMK that should be used for the AWS KMS encryption. This argument should only be used if the key is different from the default KMS-managed DynamoDB key, `alias/aws/dynamodb`. **Note:** This attribute will _not_ be populated with the ARN of _default_ keys.

### `ttl`

* `attribute_name` - (Optional) Name of the table attribute to store the TTL timestamp in.
  Required if `enabled` is `true`, must not be set otherwise.
* `enabled` - (Optional) Whether TTL is enabled.
  Default value is `false`.

### `warm_throughput`

~> **Note:** Explicitly configuring both `read_units_per_second` and `write_units_per_second` to the default/minimum values will cause Terraform to report differences.

* `read_units_per_second` - (Optional) Number of read operations a table or index can instantaneously support. For the base table, decreasing this value will force a new resource. For a global secondary index, this value can be increased or decreased without recreation. Minimum value of `12000` (default).
* `write_units_per_second` - (Optional) Number of write operations a table or index can instantaneously support. For the base table, decreasing this value will force a new resource. For a global secondary index, this value can be increased or decreased without recreation. Minimum value of `4000` (default).

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the table
* `id` - Name of the table
* `replica.*.arn` - ARN of the replica
* `replica.*.stream_arn` - ARN of the replica Table Stream. Only available when `stream_enabled = true`.
* `replica.*.stream_label` - Timestamp, in ISO 8601 format, for the replica stream. Note that this timestamp is not a unique identifier for the stream on its own. However, the combination of AWS customer ID, table name and this field is guaranteed to be unique. It can be used for creating CloudWatch Alarms. Only available when `stream_enabled = true`.
* `stream_arn` - ARN of the Table Stream. Only available when `stream_enabled = true`
* `stream_label` - Timestamp, in ISO 8601 format, for this stream. Note that this timestamp is not a unique identifier for the stream on its own. However, the combination of AWS customer ID, table name and this field is guaranteed to be unique. It can be used for creating CloudWatch Alarms. Only available when `stream_enabled = true`.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Timeouts

~> **Note:** There are a variety of default timeouts set internally. If you set a shorter custom timeout than one of the defaults, the custom timeout will not be respected as the longer of the custom or internal default will be used.

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `30m`)
* `update` - (Default `60m`)
* `delete` - (Default `10m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import DynamoDB tables using the `name`. For example:

```terraform
import {
  to = aws_dynamodb_table.basic-dynamodb-table
  id = "GameScores"
}
```

Using `terraform import`, import DynamoDB tables using the `name`. For example:

```console
% terraform import aws_dynamodb_table.basic-dynamodb-table GameScores
```
