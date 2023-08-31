---
subcategory: "Service Catalog"
layout: "aws"
page_title: "AWS: aws_servicecatalog_portfolio"
description: |-
  Provides a resource to create a Service Catalog portfolio
---

# Resource: aws_servicecatalog_portfolio

Provides a resource to create a Service Catalog Portfolio.

## Example Usage

```terraform
resource "aws_servicecatalog_portfolio" "portfolio" {
  name          = "My App Portfolio"
  description   = "List of my organizations apps"
  provider_name = "Brett"
}
```

## Argument Reference

This resource supports the following arguments:

* `name` - (Required) The name of the portfolio.
* `description` - (Required) Description of the portfolio
* `provider_name` - (Required) Name of the person or organization who owns the portfolio.
* `tags` - (Optional) Tags to apply to the connection. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - The ID of the Service Catalog Portfolio.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

- `create` - (Default `30m`)
- `read` - (Default `10m`)
- `update` - (Default `30m`)
- `delete` - (Default `30m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Service Catalog Portfolios using the Service Catalog Portfolio `id`. For example:

```terraform
import {
  to = aws_servicecatalog_portfolio.testfolio
  id = "port-12344321"
}
```

Using `terraform import`, import Service Catalog Portfolios using the Service Catalog Portfolio `id`. For example:

```console
% terraform import aws_servicecatalog_portfolio.testfolio port-12344321
```
