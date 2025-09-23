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
resource "aws_lakeformation_lf_tag" "example" {
  key    = "example"
  values = ["value"]
}

resource "aws_lakeformation_lf_tag_expression" "example" {
  name = "example"

  expression {
    tag_key    = aws_lakeformation_lf_tag.example.key
    tag_values = aws_lakeformation_lf_tag.example.values
  }
}

```

## Argument Reference

The following arguments are required:

* `name` - (Required) Name of the LF-Tag Expression.
* `expression` - (Required) A list of LF-Tag conditions (key-value pairs). See [expression](#expression) for more details.

### expression

* `tag_key` - (Required) The key-name for the LF-Tag.
* `tag_values` - (Required) A list of possible values for the LF-Tag

The following arguments are optional:

* `catalog_id` - (Optional) ID of the Data Catalog. Defaults to the account ID if not specified.
* `description` - (Optional) Description of the LF-Tag Expression.

## Attribute Reference

This resource exports no additional attributes.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Lake Formation LF Tag Expression using the `name,catalog_id`. For example:

```terraform
import {
  to = aws_lakeformation_lf_tag_expression.example
  id = "example-tag-expression,123456789012"
}
```

Using `terraform import`, import Lake Formation LF Tag Expression using the `name,catalog_id`. For example:

```console
% terraform import aws_lakeformation_lf_tag_expression.example example-tag-expression,123456789012
```
