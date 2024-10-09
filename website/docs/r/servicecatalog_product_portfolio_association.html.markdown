---
subcategory: "Service Catalog"
layout: "aws"
page_title: "AWS: aws_servicecatalog_product_portfolio_association"
description: |-
  Manages a Service Catalog Product Portfolio Association
---

# Resource: aws_servicecatalog_product_portfolio_association

Manages a Service Catalog Product Portfolio Association.

## Example Usage

### Basic Usage

```terraform
resource "aws_servicecatalog_product_portfolio_association" "example" {
  portfolio_id = "port-68656c6c6f"
  product_id   = "prod-dnigbtea24ste"
}
```

## Argument Reference

The following arguments are required:

* `portfolio_id` - (Required) Portfolio identifier.
* `product_id` - (Required) Product identifier.

The following arguments are optional:

* `accept_language` - (Optional) Language code. Valid values: `en` (English), `jp` (Japanese), `zh` (Chinese). Default value is `en`.
* `source_portfolio_id` - (Optional) Identifier of the source portfolio.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - Identifier of the association.

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

- `create` - (Default `3m`)
- `read` - (Default `10m`)
- `delete` - (Default `3m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import `aws_servicecatalog_product_portfolio_association` using the accept language, portfolio ID, and product ID. For example:

```terraform
import {
  to = aws_servicecatalog_product_portfolio_association.example
  id = "en:port-68656c6c6f:prod-dnigbtea24ste"
}
```

Using `terraform import`, import `aws_servicecatalog_product_portfolio_association` using the accept language, portfolio ID, and product ID. For example:

```console
% terraform import aws_servicecatalog_product_portfolio_association.example en:port-68656c6c6f:prod-dnigbtea24ste
```
