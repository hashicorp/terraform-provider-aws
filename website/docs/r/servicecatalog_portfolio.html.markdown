---
subcategory: "Service Catalog"
layout: "aws"
page_title: "AWS: aws_servicecatalog_portfolio"
description: |-
  Provides a resource to create a Service Catalog portfolio
---

# Resource: aws_servicecatalog_portfolio

Provides a resource to create a Service Catalog Portfolio.

Note that the portfolio cannot be used by any users unless there is 
an `aws_servicecatalog_portfolio_principal_association` explicitly granting access,
and it will be empty until at least one `aws_servicecatalog_portfolio_product_association` is created.  


## Example Usage

```hcl
resource "aws_servicecatalog_portfolio" "portfolio" {
  name          = "My App Portfolio"
  description   = "List of my organizations apps"
  provider_name = "Brett"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the portfolio.
* `description` - (Required) Description of the portfolio
* `provider_name` - (Required) Name of the person or organization who owns the portfolio.
* `tags` - (Optional) Tags to apply to the connection.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The ID of the Service Catalog Portfolio.

## Import

Service Catalog Portfolios can be imported using the `service catalog portfolio id`, e.g.

```
$ terraform import aws_servicecatalog_portfolio.testfolio port-12344321
```
