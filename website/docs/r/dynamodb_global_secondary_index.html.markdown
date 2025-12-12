---
subcategory: "DynamoDB"
layout: "aws"
page_title: "AWS: aws_dynamodb_global_secondary_index"
description: |-
  Provides a DynamoDB Global Secondary Index resource
---

# Resource: aws_dynamodb_global_secondary_index

## Example Usage

The following **experimental** DynamoDB table description models the table and GSI shown in the [AWS SDK example documentation](https://docs.aws.amazon.com/amazondynamodb/latest/developerguide/GSI.html)

```terraform
resource "aws_dynamodb_global_secondary_index" "GameTitleIndex" {
  table              = aws_dynamodb_table.basic-dynamodb-table.name
  name               = "GameTitleIndex"
  write_capacity     = 10
  read_capacity      = 10
  projection_type    = "INCLUDE"
  non_key_attributes = ["UserId"]

  key_schema {
    attribute_name = "GameTitle"
    attribute_type = "S"
    key_type       = "HASH"
  }

  key_schema {
    attribute_name = "TopScore"
    attribute_type = "N"
    key_type       = "RANGE"
  }
}

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

  ttl {
    attribute_name = "TimeToExist"
    enabled        = true
  }

  tags = {
    Name        = "dynamodb-table-1"
    Environment = "production"
  }
}
```

## Argument Reference

The following arguments are required:

* `index_name` - (Required) Name of the index.
* `key_schema` - (Required) Set of nested attribute definitions. At least 1 element defining a `HASH` is required, See below.
* `projection_type` - (Required) One of `ALL`, `INCLUDE` or `KEYS_ONLY` where `ALL` projects every attribute into the index, `KEYS_ONLY` projects  into the index only the table and index hash_key and sort_key attributes ,  `INCLUDE` projects into the index all of the attributes that are defined in `non_key_attributes` in addition to the attributes that that`KEYS_ONLY` project.
* `table_name` - (Required) Name of the table this index belongs to

The following arguments are optional:

* `non_key_attributes` - (Optional) Only required with `INCLUDE` as a projection type; a list of attributes to project into the index. These do not need to be defined as attributes on the table.
* `on_demand_throughput` - (Optional) Sets the maximum number of read and write units for the specified on-demand index. See below.
* `read_capacity` - (Optional) Number of read units for this index. Must be set if billing_mode is set to PROVISIONED.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `warm_throughput` - (Optional) Sets the number of warm read and write units for this index. See below.
* `write_capacity` - (Optional) Number of write units for this index. Must be set if billing_mode is set to PROVISIONED.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the GSI

### `key_schema`

* `attribute_name` - (Required) Name of the attribute
* `attribute_type` - (Required) Type of the attribute in the index; Valid values are `S` (string), `N` (number), `B` (binary).
* `key_type` - (Required) Key type. Valid values are `HASH`, `RANGE`.

## Migrating

For each block `global_secondary_index` create a new `aws_dynamodb_global_secondary_index` resource with the same configuration as the block you're replacing and add the following line into the new resource:

```
    # see Example section; replace $resource$ with the actual resource name
    table_name = aws_dynamodb_global_secondary_index.$resource$.name
```

Using `terraform import`, import DynamoDB global secondary indexes using the `arn`. For example:

```console
% terraform import aws_dynamodb_global_secondary_index.GameScores arn:aws:dynamodb:eu-west-1:123456789012:table/GameScores/index/GameTitleIndex
```

Run `terraform plan` to validate.

~> **Note:** You can use either `global_secondary_index` blocks or `aws_dynamodb_global_secondary_index` resources. You cannot use both on the same table. If you chose to migrate to this new resource, you must migrate all the global secondary indexes that the table defines.

## Reverting

For each `aws_dynamodb_global_secondary_index` that you want to remove, you must create a `global_secondary_index` inside the table where it belongs.

Detach the `aws_dynamodb_global_secondary_index` resource from state with:

```
% terraform state rm aws_dynamodb_global_secondary_index.$resource$
```

Run `terraform plan` to validate.
