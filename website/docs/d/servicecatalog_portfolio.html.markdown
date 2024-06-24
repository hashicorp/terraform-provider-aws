---
subcategory: "Service Catalog"
layout: "aws"
page_title: "AWS: aws_servicecatalog_portfolio"
description: |-
  Provides information for a Service Catalog Portfolio.
---

# Data Source: aws_servicecatalog_portfolio

Provides information for a Service Catalog Portfolio.

## Example Usage

```terraform
data "aws_servicecatalog_portfolio" "portfolio" {
  id = "port-07052002"
}
```

## Argument Reference

The following arguments are required:

* `id` - (Required) Portfolio identifier.

The following arguments are optional:

* `accept_language` - (Optional) Language code. Valid values: `en` (English), `jp` (Japanese), `zh` (Chinese). Default value is `en`.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `arn` - Portfolio ARN.
* `created_time` - Time the portfolio was created.
* `description` - Description of the portfolio
* `name` - Portfolio name.
* `provider_name` - Name of the person or organization who owns the portfolio.
* `tags` - Tags applied to the portfolio.
