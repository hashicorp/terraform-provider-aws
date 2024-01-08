---
subcategory: "Service Catalog"
layout: "aws"
page_title: "AWS: aws_servicecatalog_constraint"
description: |-
  Provides information on a Service Catalog Constraint
---

# Data Source: aws_servicecatalog_constraint

Provides information on a Service Catalog Constraint.

## Example Usage

### Basic Usage

```terraform
data "aws_servicecatalog_constraint" "example" {
  accept_language = "en"
  id              = "cons-hrvy0335"
}
```

## Argument Reference

The following arguments are required:

* `id` - Constraint identifier.

The following arguments are optional:

* `accept_language` - (Optional) Language code. Valid values: `en` (English), `jp` (Japanese), `zh` (Chinese). Default value is `en`.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `description` - Description of the constraint.
* `owner` - Owner of the constraint.
* `parameters` - Constraint parameters in JSON format.
* `portfolio_id` - Portfolio identifier.
* `product_id` - Product identifier.
* `status` - Constraint status.
* `type` - Type of constraint. Valid values are `LAUNCH`, `NOTIFICATION`, `RESOURCE_UPDATE`, `STACKSET`, and `TEMPLATE`.
