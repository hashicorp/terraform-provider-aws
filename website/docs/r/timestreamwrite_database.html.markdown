---
subcategory: "Timestream Write"
layout: "aws"
page_title: "AWS: aws_timestreamwrite_database"
description: |-
  Provides a Timestream database resource.
---

# Resource: aws_timestreamwrite_database

Provides a Timestream database resource.

## Example Usage

### Basic usage

```hcl
resource "aws_timestreamwrite_database" "test_database" {
  database_name = "database-example"
}
```

### Full usage

```hcl
resource "aws_timestreamwrite_database" "test_database" {
  database_name = "database-example"
  kms_key_id    = aws_kms_key.foo.arn

  tags = {
    Name = "value"
  }
}
```

## Argument Reference

The following arguments are supported:

* `database_name` – (Required) The name of the Timestream database. Minimum length of 3. Maximum length of 64.
* `kms_key_id` - (Optional) The KMS key for the database. You can specify a key ARN.
    If the KMS key is not specified, the database will be encrypted with a Timestream managed KMS key located in your account. Refer to [AWS managed KMS keys](https://docs.aws.amazon.com/kms/latest/developerguide/concepts.html#aws-managed-cmk) for more info.
* `tags` - (Optional) A map of tags to assign to the resource


## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - The Amazon Resource Name that uniquely identifies this database.
* `kms_key_id` - The identifier of the KMS key used to encrypt the data stored in the database.

## Import

Timestream databases can be imported using the `database_name`, e.g.

```
$ terraform import aws_timestreamwrite_database.my_database my_database
```

