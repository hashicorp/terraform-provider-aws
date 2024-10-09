---
subcategory: "Glue"
layout: "aws"
page_title: "AWS: aws_glue_user_defined_function"
description: |-
  Provides a Glue User Defined Function.
---

# Resource: aws_glue_user_defined_function

Provides a Glue User Defined Function Resource.

## Example Usage

```terraform
resource "aws_glue_catalog_database" "example" {
  name = "my_database"
}

resource "aws_glue_user_defined_function" "example" {
  name          = "my_func"
  catalog_id    = aws_glue_catalog_database.example.catalog_id
  database_name = aws_glue_catalog_database.example.name
  class_name    = "class"
  owner_name    = "owner"
  owner_type    = "GROUP"

  resource_uris {
    resource_type = "ARCHIVE"
    uri           = "uri"
  }
}
```

## Argument Reference

This resource supports the following arguments:

* `name` - (Required) The name of the function.
* `catalog_id` - (Optional) ID of the Glue Catalog to create the function in. If omitted, this defaults to the AWS Account ID.
* `database_name` - (Required) The name of the Database to create the Function.
* `class_name` - (Required) The Java class that contains the function code.
* `owner_name` - (Required) The owner of the function.
* `owner_type` - (Required) The owner type. can be one of `USER`, `ROLE`, and `GROUP`.
* `resource_uris` - (Optional) The configuration block for Resource URIs. See [resource uris](#resource-uris) below for more details.

### Resource URIs

* `resource_type` - (Required) The type of the resource. can be one of `JAR`, `FILE`, and `ARCHIVE`.
* `uri` - (Required) The URI for accessing the resource.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id`- The id of the Glue User Defined Function.
* `arn`- The ARN of the Glue User Defined Function.
* `create_time`- The time at which the function was created.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Glue User Defined Functions using the `catalog_id:database_name:function_name`. If you have not set a Catalog ID specify the AWS Account ID that the database is in. For example:

```terraform
import {
  to = aws_glue_user_defined_function.func
  id = "123456789012:my_database:my_func"
}
```

Using `terraform import`, import Glue User Defined Functions using the `catalog_id:database_name:function_name`. If you have not set a Catalog ID specify the AWS Account ID that the database is in. For example:

```console
% terraform import aws_glue_user_defined_function.func 123456789012:my_database:my_func
```
