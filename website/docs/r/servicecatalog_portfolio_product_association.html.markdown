---
subcategory: "Service Catalog"
layout: "aws"
page_title: "AWS: aws_servicecatalog_portfolio_product_association"
description: |-
  Provides a resource to control the association of a Service Catalog Product with a Portfolio
---

# aws_servicecatalog_portfolio_product_association

Provides a resource to control the association of a Service Catalog Product with a Portfolio.

This is necessary for a product to be provisioned, as it must be in at least one portfolio.

In most cases this is simple and straightforward. 
However, there are some pathological edge cases that can arise 
as the association is not an identifiable resource in the usual sense. 
For instance if an association between a given portfolio and product were created twice,
with two instances in Terraform, and then one of them is deleted, 
there would be one instance remaining in Terraform but
the association be absent in AWS until the remaining Terraform instance is re-applied.   


## Example Usage

```hcl
resource "aws_servicecatalog_portfolio_product_association" "test" {
  portfolio_id = "port-01234567890abc"
  product_id   = "prod-abcdefghijklm"
}
```

## Argument Reference

The following arguments are supported:

* `portfolio_id` - (Required) The portfolio identifier
* `product_id` - (Required) The product identifier


## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - A generated ID to represent this association, of the form `${portfolio_id}:${product_id}`.


## Import

Service Catalog Portfolio-Product Associations can be imported using the ID constructed 
from the portfolio and product ids, e.g.

```
$ terraform import aws_servicecatalog_portfolio_product_association.test port-01234567890abc:prod-abcdefghijklm
```
