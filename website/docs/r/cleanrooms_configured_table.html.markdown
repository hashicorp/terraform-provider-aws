---
subcategory: "Clean Rooms"
layout: "aws"
page_title: "AWS: aws_cleanrooms_configured_table"
description: |-
  Provides a Clean Rooms Configured Table.
---

# Resource: aws_cleanrooms_configured_table

Provides a AWS Clean Rooms configured table. Configured tables are used to represent references to existing tables in the AWS Glue Data Catalog.

## Example Usage

### Configured table with tags

```terraform
resource "aws_cleanrooms_configured_table" "test_configured_table" {
  name            = "terraform-example-table"
  description     = "I made this table with terraform!"
  analysis_method = "DIRECT_QUERY"
  allowed_columns = [
    "column1",
    "column2",
    "column3",
  ]

  table_reference {
    database_name = "example_database"
    table_name    = "example_table"
  }

  tags = {
    Project = "Terraform"
  }
}
```

## Argument Reference

This resource supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `name` - (Required) - The name of the configured table.
* `description` - (Optional) - A description for the configured table.
* `analysis_method` - (Required) - The analysis method for the configured table. The only valid value is currently `DIRECT_QUERY`.
* `allowed_columns` - (Required - Forces new resource) - The columns of the references table which will be included in the configured table.
* `table_reference` - (Required - Forces new resource) - A reference to the AWS Glue table which will be used to create the configured table.
* `table_reference.database_name` - (Required - Forces new resource) - The name of the AWS Glue database which contains the table.
* `table_reference.table_name` - (Required - Forces new resource) - The name of the AWS Glue table which will be used to create the configured table.
* `tags` - (Optional) - Key value pairs which tag the configured table.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - The ARN of the configured table.
* `id` - The ID of the configured table.
* `create_time` - The date and time the configured table was created.
* `update_time` - The date and time the configured table was last updated.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

- `create` - (Default `1m`)
- `update` - (Default `1m`)
- `delete` - (Default `1m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import `aws_cleanrooms_configured_table` using the `id`. For example:

```terraform
import {
  to = aws_cleanrooms_configured_table.table
  id = "1234abcd-12ab-34cd-56ef-1234567890ab"
}
```

Using `terraform import`, import `aws_cleanrooms_configured_table` using the `id`. For example:

```console
% terraform import aws_cleanrooms_configured_table.table 1234abcd-12ab-34cd-56ef-1234567890ab
```
