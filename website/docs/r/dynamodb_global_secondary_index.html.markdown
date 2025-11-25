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

resource "aws_dynamodb_global_secondary_index" "GameTitleIndex" {
  table              = aws_dynamodb_table.basic-dynamodb-table.name
  name               = "GameTitleIndex"
  hash_key           = "GameTitle"
  range_key          = "TopScore"
  range_key_type     = "N"
  write_capacity     = 10
  read_capacity      = 10
  projection_type    = "INCLUDE"
  non_key_attributes = ["UserId"]
}
```

## Argument Reference

The following arguments are required:

* `hash_key` - (Required) Name of the hash key in the index
* `name` - (Required) Name of the index.
* `projection_type` - (Required) One of `ALL`, `INCLUDE` or `KEYS_ONLY` where `ALL` projects every attribute into the index, `KEYS_ONLY` projects  into the index only the table and index hash_key and sort_key attributes ,  `INCLUDE` projects into the index all of the attributes that are defined in `non_key_attributes` in addition to the attributes that that`KEYS_ONLY` project.
* `table` - (Required) Name of the table this index belongs to

The following arguments are optional:

* `hash_key_type` - (Optional) Type of the hash key in the index; One of `S`, `N` or `B`; If the attribute exists in the table, the value will be inferred
* `non_key_attributes` - (Optional) Only required with `INCLUDE` as a projection type; a list of attributes to project into the index. These do not need to be defined as attributes on the table.
* `on_demand_throughput` - (Optional) Sets the maximum number of read and write units for the specified on-demand index. See below.
* `range_key` - (Optional) Name of the range key; must be defined
* `range_key_type` - (Optional) Type of the range key; One of `S`, `N` or `B`; If the attribute exists in the table, the value will be inferred
* `read_capacity` - (Optional) Number of read units for this index. Must be set if billing_mode is set to PROVISIONED.
* `warm_throughput` - (Optional) Sets the number of warm read and write units for this index. See below.
* `write_capacity` - (Optional) Number of write units for this index. Must be set if billing_mode is set to PROVISIONED.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - name of the GSI

## Migrating

For each block `global_secondary_index` create a new `aws_dynamodb_global_secondary_index` resource with the same configuration as the block you're replacing and add the following line into the new resource:

```
    # see Example section; replace $resource$ with the actual resource name
    table = aws_dynamodb_global_secondary_index.$resource$.name
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
% terraform state rm aws_dynamodb_global_secondary_index.$resouce$
```

Run `terraform plan` to validate.
