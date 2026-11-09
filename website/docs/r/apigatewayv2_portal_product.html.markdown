---
subcategory: "API Gateway V2"
layout: "aws"
page_title: "AWS: aws_apigatewayv2_portal_product"
description: |-
  Manages an AWS API Gateway V2 Portal Product.
---

# Resource: aws_apigatewayv2_portal_product

Manages an AWS API Gateway V2 Portal Product.

A portal product is a business-focused grouping of REST API endpoints that is surfaced in an API Gateway developer portal. Portal products are created independently of any portal; a portal then references them by ARN.

## Example Usage

### Basic Usage

```terraform
resource "aws_apigatewayv2_portal_product" "example" {
  display_name = "AdoptAnimals"
  description  = "APIs for browsing and adopting shelter animals."
}
```

### Multiple Products

```terraform
resource "aws_apigatewayv2_portal_product" "adopt" {
  display_name = "AdoptAnimals"
  description  = "APIs for browsing and adopting shelter animals."

  tags = {
    Audience = "public"
  }
}

resource "aws_apigatewayv2_portal_product" "veterinary" {
  display_name = "VeterinaryRecords"
  description  = "APIs for managing animal medical records."

  tags = {
    Audience = "partner"
  }
}
```

## Argument Reference

The following arguments are required:

* `display_name` - (Required) Name of the portal product as it appears in a published portal. Must be between 1 and 255 characters in length.

The following arguments are optional:

* `description` - (Optional) Description of the portal product. Must be at most 1024 characters in length. To clear an existing description, set this to an empty string; removing the argument leaves the current description in place.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `tags` - (Optional) Map of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level. Tag values must be at least one character long; API Gateway rejects empty and null tag values.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `last_modified` - Timestamp when the portal product was last modified, in RFC3339 format.
* `portal_product_arn` - ARN of the portal product.
* `portal_product_id` - Unique identifier of the portal product.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Import

In Terraform v1.12.0 and later, the [`import` block](https://developer.hashicorp.com/terraform/language/import) can be used with the `identity` attribute. For example:

```terraform
import {
  to = aws_apigatewayv2_portal_product.example
  identity = {
    portal_product_id = "abcdef1234"
  }
}

resource "aws_apigatewayv2_portal_product" "example" {
  ### Configuration omitted for brevity ###
}
```

### Identity Schema

#### Required

* `portal_product_id` (String) ID of the portal product.

#### Optional

* `account_id` (String) AWS Account where this resource is managed.
* `region` (String) Region where this resource is managed.

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import API Gateway V2 Portal Products using `portal_product_id`. For example:

```terraform
import {
  to = aws_apigatewayv2_portal_product.example
  id = "abcdef1234"
}
```

Using `terraform import`, import API Gateway V2 Portal Products using `portal_product_id`. For example:

```console
% terraform import aws_apigatewayv2_portal_product.example abcdef1234
```
