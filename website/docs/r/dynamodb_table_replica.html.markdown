---
subcategory: "DynamoDB"
layout: "aws"
page_title: "AWS: aws_dynamodb_table_replica"
description: |-
  Provides a DynamoDB table replica resource
---

# Resource: aws_dynamodb_table_replica

Provides a DynamoDB table replica resource for [DynamoDB Global Tables V2 (version 2019.11.21)](https://docs.aws.amazon.com/amazondynamodb/latest/developerguide/globaltables.V2.html).

~> **Note:** Use `lifecycle` [`ignore_changes`](https://www.terraform.io/docs/configuration/meta-arguments/lifecycle.html#ignore_changes) for `replica` in the associated [aws_dynamodb_table](/docs/providers/aws/r/dynamodb_table.html) configuration.

~> **Note:** Do not use the `replica` configuration block of [aws_dynamodb_table](/docs/providers/aws/r/dynamodb_table.html) together with this resource as the two configuration options are mutually exclusive.

## Example Usage

### Basic Example

```terraform
provider "aws" {
  alias  = "main"
  region = "us-west-2"
}

provider "aws" {
  alias  = "alt"
  region = "us-east-2"
}

resource "aws_dynamodb_table" "example" {
  provider         = "aws.main"
  name             = "TestTable"
  hash_key         = "BrodoBaggins"
  billing_mode     = "PAY_PER_REQUEST"
  stream_enabled   = true
  stream_view_type = "NEW_AND_OLD_IMAGES"

  attribute {
    name = "BrodoBaggins"
    type = "S"
  }

  lifecycle {
    ignore_changes = [replica]
  }
}

resource "aws_dynamodb_table_replica" "example" {
  provider         = "aws.alt"
  global_table_arn = aws_dynamodb_table.example.arn

  tags = {
    Name = "IZPAWS"
    Pozo = "Amargo"
  }
}
```

## Argument Reference

Required arguments:

* `global_table_arn` - (Required) ARN of the _main_ or global table which this resource will replicate.

Optional arguments:

* `kms_key_arn` - (Optional, Forces new resource) ARN of the CMK that should be used for the AWS KMS encryption. This argument should only be used if the key is different from the default KMS-managed DynamoDB key, `alias/aws/dynamodb`. **Note:** This attribute will _not_ be populated with the ARN of _default_ keys.
* `point_in_time_recovery` - (Optional) Whether to enable Point In Time Recovery for the replica. Default is `false`.
* `table_class_override` - (Optional, Forces new resource) Storage class of the table replica. Valid values are `STANDARD` and `STANDARD_INFREQUENT_ACCESS`. If not used, the table replica will use the same class as the global table.
* `tags` - (Optional) Map of tags to populate on the created table. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - ARN of the table replica.
* `id` - Name of the table and region of the main global table joined with a semicolon (_e.g._, `TableName:us-east-1`).
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `30m`)
* `update` - (Default `30m`)
* `delete` - (Default `20m`)

## Import

DynamoDB table replicas can be imported using the `table-name:main-region`, _e.g._,

~> **Note:** When importing, use the region where the initial or _main_ global table resides, _not_ the region of the replica.

```
$ terraform import aws_dynamodb_table_replica.example TestTable:us-west-2
```
