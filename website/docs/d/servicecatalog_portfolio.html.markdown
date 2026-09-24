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

```terraform
data "aws_servicecatalog_portfolio" "portfolio" {
  name = "example-portfolio"
}
```

## Argument Reference

Exactly one of `id` or `name` must be specified.

The following arguments are optional:

* `id` - (Optional) Portfolio identifier.
* `name` - (Optional) Portfolio name. Returns an error if more than one portfolio matches the given name.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `accept_language` - (Optional) Language code. Valid values: `en` (English), `jp` (Japanese), `zh` (Chinese). Default value is `en`.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `arn` - Portfolio ARN.
* `created_time` - Time the portfolio was created.
* `description` - Description of the portfolio
* `provider_name` - Name of the person or organization who owns the portfolio.
* `tags` - Tags applied to the portfolio.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `read` - (Default `10m`)
