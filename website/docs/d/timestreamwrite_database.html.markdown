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
data "aws_timestreamwrite_database" "test" {
  name = "database-example"
}
```

## Argument Reference

This data source supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `database_name` - (Required) The name of the Timestream database. Minimum length of 3. Maximum length of 256.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `arn` - The ARN that uniquely identifies this database.
* `created_time` - Creation time of database.
* `database_name` - (Required) The name of the Timestream database. Minimum length of 3. Maximum length of 256.
* `kms_key_id` - The ARN of the KMS key used to encrypt the data stored in the database.
* `last_updated_time` - Last time database was updated.
* `table_count` -  Total number of tables in the Timestream database.
