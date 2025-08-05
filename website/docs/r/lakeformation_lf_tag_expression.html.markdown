---
subcategory: "Lake Formation"
layout: "aws"
page_title: "AWS: aws_lakeformation_lf_tag_expression"
description: |-
  Terraform resource for managing an AWS Lake Formation LF Tag Expression.
---
# Resource: aws_lakeformation_lf_tag_expression

Terraform resource for managing an AWS Lake Formation LF Tag Expression.

## Example Usage

### Basic Usage

```terraform
resource "aws_lakeformation_lf_tag_expression" "example" {
  name = "example-tag-expression"
  
  tag_expression = {
    "Environment" = ["dev", "staging", "prod"]
    "Department"  = ["engineering", "marketing"]
  }
  
  description = "Example LF Tag Expression for demo purposes"
}
```

## Argument Reference

The following arguments are required:

* `name` - (Required) Name of the LF-Tag Expression.
* `tag_expression` - (Required) Mapping of tag keys to lists of allowed values.

The following arguments are optional:

* `catalog_id` - (Optional) ID of the Data Catalog. Defaults to the account ID if not specified.
* `description` - (Optional) Description of the LF-Tag Expression.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - Primary identifier (catalog_id:name).

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `30m`)
* `update` - (Default `30m`)
* `delete` - (Default `30m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Lake Formation LF Tag Expression using the `catalog_id:name`. For example:

```terraform
import {
  to = aws_lakeformation_lf_tag_expression.example
  id = "123456789012:example-tag-expression"
}
```

Using `terraform import`, import Lake Formation LF Tag Expression using the `catalog_id:name`. For example:

```console
% terraform import aws_lakeformation_lf_tag_expression.example 123456789012:example-tag-expression
```
