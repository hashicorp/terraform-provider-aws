---
subcategory: "Service Catalog AppRegistry"
layout: "aws"
page_title: "AWS: aws_servicecatalogappregistry_application_attribute_group_association"
description: |-
  Terraform resource for managing an AWS Service Catalog AppRegistry Application Attribute Group Association.
---
# Resource: aws_servicecatalogappregistry_application_attribute_group_association

Terraform resource for managing an AWS Service Catalog AppRegistry Application Attribute Group Association.

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

resource "aws_servicecatalogappregistry_application_attribute_group_association" "example" {
  application_id     = aws_servicecatalogappregistry_application.example.id
  attribute_group_id = aws_servicecatalogappregistry_attribute_group.example.id
}
```

## Argument Reference

The following arguments are required:

* `application_id` - (Required) ID of the Application to associate
* `attribute_group_id` - (Required) ID of the Attribute Group to associate

The following arguments are optional:

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Service Catalog AppRegistry Application Attribute Group Association using the `example_id_arg`. For example:

```terraform
import {
  to = aws_servicecatalogappregistry_application_attribute_group_association.example
  id = "12456778723424sdffsdfsdq34,12234t3564dsfsdf34asff4ww3"
}
```

Using `terraform import`, import Service Catalog AppRegistry Application Attribute Group Association using the `12456778723424sdffsdfsdq34,12234t3564dsfsdf34asff4ww3`. For example:

```console
% terraform import aws_servicecatalogappregistry_application_attribute_group_association.example 12456778723424sdffsdfsdq34,12234t3564dsfsdf34asff4ww3
```
