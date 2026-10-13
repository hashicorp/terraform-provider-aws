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

* `catalog_id` - (Optional) ID of the Glue Catalog to create the function in. If omitted, this defaults to the AWS Account ID.
* `class_name` - (Required) Java class that contains the function code.
* `database_name` - (Required) Name of the Database to create the Function.
* `name` - (Required) Name of the function.
* `owner_name` - (Required) Owner of the function.
* `owner_type` - (Required) Owner type. can be one of `USER`, `ROLE`, and `GROUP`.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `resource_uris` - (Optional) Configuration block for Resource URIs. See [`resource_uris` Block](#resource_uris-block) below for more details.

### `resource_uris` Block

* `resource_type` - (Required) Type of the resource. can be one of `JAR`, `FILE`, and `ARCHIVE`.
* `uri` - (Required) URI for accessing the resource.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn`- ARN of the Glue User Defined Function.
* `create_time`- Time at which the function was created.
* `id`- ID of the Glue User Defined Function.

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
