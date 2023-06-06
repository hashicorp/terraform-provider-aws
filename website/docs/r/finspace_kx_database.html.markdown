---
subcategory: "FinSpace"
layout: "aws"
page_title: "AWS: aws_finspace_kx_database"
description: |-
  Terraform resource for managing an AWS FinSpace Kx Database.
---

# Resource: aws_finspace_kx_database

Terraform resource for managing an AWS FinSpace Kx Database.

## Example Usage

### Basic Usage

```terraform
resource "aws_kms_key" "example" {
  description             = "Example KMS Key"
  deletion_window_in_days = 7
}

resource "aws_finspace_kx_environment" "example" {
  name     = "my-tf-kx-environment"
  kms_key_id =aws_kms_key.example.arn
}

resource "aws_finspace_kx_database" "example" {
  environment_id = aws_finspace_kx_environment.example.id
  name = "my-tf-kx-database"
  description = "Example database description"
}
```

## Argument Reference

The following arguments are required:

* `name` - (Required) Name of the KX database.
* `environment_id` - (Required) Unique identifier for the KX environment.

The following arguments are optional:

* `description` - (Optional) Description of the KX database.
* `tags` - (Optional) List of key-value pairs to label the KX database.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - Amazon Resource Name (ARN) identifier of the KX database.
* `created_timestamp` - Timestamp at which the databse is created in FinSpace. Value determined as epoch time in seconds. For example, the value for Monday, November 1, 2021 12:00:00 PM UTC is specified as 1635768000.
* `last_modified_timestamp` - Last timestamp at which the database was updated in FinSpace. Value determined as epoch time in seconds. For example, the value for Monday, November 1, 2021 12:00:00 PM UTC is specified as 1635768000.
* `tags_all` - Map of tags assigned to the resource.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `30m`)
* `update` - (Default `30m`)
* `delete` - (Default `30m`)
