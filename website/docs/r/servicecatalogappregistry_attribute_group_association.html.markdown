---
subcategory: "Service Catalog AppRegistry"
layout: "aws"
page_title: "AWS: aws_servicecatalogappregistry_attribute_group_association"
description: |-
  Terraform resource for managing an AWS Service Catalog AppRegistry Attribute Group Association.
---
# Resource: aws_servicecatalogappregistry_attribute_group_association

Terraform resource for managing an AWS Service Catalog AppRegistry Attribute Group Association.

## Example Usage

### Basic Usage

```terraform
resource "aws_servicecatalogappregistry_application" "example" {
  name = "example-app"
}

resource "aws_servicecatalogappregistry_attribute_group" "example" {
  name        = "example"
  description = "example description"

  attributes = jsonencode({
    app   = "exampleapp"
    group = "examplegroup"
  })
}

resource "aws_servicecatalogappregistry_attribute_group_association" "example" {
  application_id     = aws_servicecatalogappregistry_application.example.id
  attribute_group_id = aws_servicecatalogappregistry_attribute_group.example.id
}
```

## Argument Reference

The following arguments are required:

* `application_id` - (Required) ID of the application.
* `attribute_group_id` - (Required) ID of the attribute group to associate with the application.

## Attribute Reference

This resource exports no additional attributes.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Service Catalog AppRegistry Attribute Group Association using the `application_id` and `attribute_group_id` arguments separated by a comma (`,`). For example:

```terraform
import {
  to = aws_servicecatalogappregistry_attribute_group_association.example
  id = "12456778723424sdffsdfsdq34,12234t3564dsfsdf34asff4ww3"
}
```

Using `terraform import`, import Service Catalog AppRegistry Attribute Group Association using `application_id` and `attribute_group_id` arguments separated by a comma (`,`). For example:

```console
% terraform import aws_servicecatalogappregistry_attribute_group_association.example 12456778723424sdffsdfsdq34,12234t3564dsfsdf34asff4ww3
```
