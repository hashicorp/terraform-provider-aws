---
subcategory: "Timestream Write"
layout: "aws"
page_title: "AWS: aws_timestreamwrite_database"
description: |-
  Terraform data source for managing an AWS Timestream Write Database.
---

# Data Source: aws_timestreamwrite_database

Terraform data source for managing an AWS Timestream Write Database.

## Example Usage

### Basic Usage

```terraform
data "aws_timestreamwrite_database" "example" {
  database_name = "database-example"
}
```

## Argument Reference

The following arguments are required:

* `database_name` – (Required) The name of the Timestream database. Minimum length of 3. Maximum length of 64.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `id` - The name of the Timestream database.
* `arn` - The ARN that uniquely identifies this database.
* `kms_key_id` - The ARN of the KMS key used to encrypt the data stored in the database.
* `table_count` - The total number of tables found within the Timestream database.
