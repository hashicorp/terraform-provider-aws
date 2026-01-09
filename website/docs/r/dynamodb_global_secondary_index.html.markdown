---
subcategory: "DynamoDB"
layout: "aws"
page_title: "AWS: aws_dynamodb_global_secondary_index"
description: |-
  Provides a DynamoDB Global Secondary Index resource
---

# Resource: aws_dynamodb_global_secondary_index

!> The resource type `aws_dynamodb_global_secondary_index` is an experimental feature. The schema or behavior may change without notice, and it is not subject to the backwards compatibility guarantee of the provider.

~> The resource type `aws_dynamodb_global_secondary_index` can be enabled by setting the environment variable `TF_AWS_EXPERIMENT_dynamodb_global_secondary_index` to any value. If not enabled, use of `aws_dynamodb_global_secondary_index` will result in an error when running Terraform.

-> Please provide feedback, positive or negative, at https://github.com/hashicorp/terraform-provider-aws/issues/45640. User feedback will determine if this experiment is a success.

!> **WARNING:** Do not combine `aws_dynamodb_global_secondary_index` resources in conjunction with `global_secondary_index` on [`aws_dynamodb_table`](dynamodb_table.html). Doing so may cause conflicts, perpertual differences, and Global Secondary Indexes being overwritten.

## Example Usage

```terraform
resource "aws_dynamodb_global_secondary_index" "example" {
  table = aws_dynamodb_table.example.name
  name  = "GameTitleIndex"

  projection {
    projection_type    = "INCLUDE"
    non_key_attributes = ["UserId"]
  }

  provisioned_throughput {
    write_capacity_units = 10
    read_capacity_units  = 10
  }

  key_schema {
    attribute_name = "GameTitle"
    attribute_type = "S"
    key_type       = "HASH"
  }
}

resource "aws_dynamodb_table" "example" {
  name           = "example"
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
}
```

## Migrating

Use the following steps to migrate existing Global Secondary Indexes defined inline in `global_secondary_index` on an `aws_dynamodb_table`.

For each block `global_secondary_index` create a new `aws_dynamodb_global_secondary_index` resource with configuration corresponding to the existing block.

For example, starting with the following configuration:

```terraform
resource "aws_dynamodb_table" "example" {
  name           = "example-table"
  hash_key       = "example-key"
  read_capacity  = 1
  write_capacity = 1

  global_secondary_index {
    name            = "example-index-1"
    projection_type = "ALL"
    hash_key        = "example-gsi-key-1"
    read_capacity   = 1
    write_capacity  = 1
  }

  global_secondary_index {
    name            = "example-index-2"
    projection_type = "ALL"
    hash_key        = "example-gsi-key-2"
    read_capacity   = 1
    write_capacity  = 1
  }

  attribute {
    name = "example-key"
    type = "S"
  }

  attribute {
    name = "example-gsi-key-1"
    type = "S"
  }

  attribute {
    name = "example-gsi-key-2"
    type = "S"
  }
}
```

Update the configuration to the following. Note that the schema of `aws_dynamodb_global_secondary_index` has some differences with `global_secondary_index` on `aws_dynamodb_table`.

If using Terraform versions prior to v1.5.0, remove the `import` blocks and use the `terraform import` command.

```terraform
import {
  to = aws_dynamodb_global_secondary_index.example1
  id = "${aws_dynamodb_table.example.name},example-index-1"
}

import {
  to = aws_dynamodb_global_secondary_index.example2
  id = "${aws_dynamodb_table.example.name},example-index-2"
}

resource "aws_dynamodb_global_secondary_index" "example1" {
  table_name = aws_dynamodb_table.test.name
  index_name = "example-index-1"
  projection {
    projection_type = "ALL"
  }

  provisioned_throughput {
    read_capacity_units  = 1
    write_capacity_units = 1
  }

  key_schema {
    attribute_name = "example-gsi-key-1"
    attribute_type = "S"
    key_type       = "HASH"
  }
}

resource "aws_dynamodb_global_secondary_index" "example2" {
  table_name = aws_dynamodb_table.test.name
  index_name = "example-index-2"
  projection {
    projection_type = "ALL"
  }

  provisioned_throughput {
    read_capacity_units  = 1
    write_capacity_units = 1
  }

  key_schema {
    attribute_name = "example-gsi-key-2"
    attribute_type = "S"
    key_type       = "HASH"
  }
}

resource "aws_dynamodb_table" "test" {
  name           = "example-table"
  hash_key       = "example-key"
  read_capacity  = 1
  write_capacity = 1

  attribute {
    name = "example-key"
    type = "S"
  }
}
```

For more details on importing `aws_dynamodb_global_secondary_index` resources, see [the Import section below](#import).

## Argument Reference

The following arguments are required:

* `index_name` - (Required) Name of the index.
* `key_schema` - (Required) Set of nested attribute definitions.
  At least 1 element defining a `HASH` is required.
  All elements with the `key_type` of `HASH` must precede elements with `key_type` of `RANGE`.
  Changing any values in `key_schema` will re-create the resource.
  See [`key_schema` below](#key_schema).
* `projection` - (Required) Describes which attributes from the table are represented in the index.
  See [`projection` below](#projection).
* `table_name` - (Required) Name of the table this index belongs to.

The following arguments are optional:

* `on_demand_throughput` - (Optional) Sets the maximum number of read and write units for the index.
  See [`on_demand_throughput` below](#on_demand_throughput).
  Only valid if the table's `billing_mode` is `PAY_PER_REQUEST`.
* `provisioned_throughput` - (Optional) Provisioned throughput for the index.
  See [`provisioned_throughput` below](#provisioned_throughput).
  Required if the table's `billing_mode` is `PROVISIONED`.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `warm_throughput` - (Optional) Sets the number of warm read and write units for this index.
  See [`warm_throughput` below](#warm_throughput).

### `key_schema`

* `attribute_name` - (Required) Name of the attribute.
* `attribute_type` - (Required) Type of the attribute in the index.
  Valid values are `S` (string), `N` (number), or `B` (binary).
* `key_type` - (Required) Key type.
  Valid values are `HASH` or `RANGE`.

### `on_demand_throughput`

* `max_read_request_units` - (Optional) Maximum number of read request units for this index.
* `max_write_request_units` - (Optional) Maximum number of write request units for this index.

### `projection`

* `non_key_attributes` - (Optional) Specifies which additional attributes to include in the index.
  Only valid when `projection_type` is `INCLUDE`.`
* `projection_type` - (Required) The set of attributes represented in the index.
  One of `ALL`, `INCLUDE`, or `KEYS_ONLY`.

### `provisioned_throughput`

* `read_capacity_units` - (Required) Number of read capacity units for this index.
* `write_capacity_units` - (Required) Number of write capacity units for this index.

### `warm_throughput`

* `read_units_per_second` - (Required) Number of read operations this index can instantaneously support.
* `write_units_per_second` - (Required) Number of write operations this index can instantaneously support.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the GSI.

## Import

In Terraform v1.12.0 and later, the [`import` block](https://developer.hashicorp.com/terraform/language/import) can be used with the `identity` attribute. For example:

```terraform
import {
  to = aws_dynamodb_global_secondary_index.example
  identity = {
    "table_name" = "example-table"
    "index_name" = "example-index"
  }
}

resource "aws_dynamodb_global_secondary_index" "example" {
  ### Configuration omitted for brevity ###
}
```

### Identity Schema

#### Required

* `index_name` (String) Name of the index.
* `table_name` (String) Name of the table this index belongs to.

#### Optional

* `account_id` (String) AWS Account where this resource is managed.
* `region` (String) Region where this resource is managed.

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import DynamoDB tables using the `table_name` and `index_name`, separated by a comma. For example:

```terraform
import {
  to = aws_dynamodb_global_secondary_index.example
  id = "example-table,example-index"
}
```

Using `terraform import`, import DynamoDB tables using the `table_name` and `index_name`, separated by a comma. For example:

```console
% terraform import aws_dynamodb_global_secondary_index.example 'example-table,example-index'
```
