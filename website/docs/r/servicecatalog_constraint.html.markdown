---
subcategory: "Service Catalog"
layout: "aws"
page_title: "AWS: aws_servicecatalog_constraint"
description: |-
  Provides a resource to control the rules that are applied to a product in a specific portfolio when the end users launches it
---

# Resource: aws_servicecatalog_constraint

Provides a resource to control the rules that are applied to a product in a specific portfolio when the end users launches it.

When the end users launches the product, they will see the rules you have applied using constraints. You can apply constraints to a product once it is put into a portfolio. Constraints are active as soon as you create them, and they're applied to all current versions of a product that have not been launched. 

## Example Usage

```hcl
resource "aws_servicecatalog_constraint" "constraint" {
    description = "Only Team Alpha Admins may launch"
  parameters = <<EOF
{
  "LocalRoleName": "teams/alpha/Admin"
}
EOF
  portfolio_id = aws_servicecatalog_portfolio.myportfolio.id
  product_id = aws_servicecatalog_product.myproduct.id
  type = "LAUNCH"
}
```

## Argument Reference

The following arguments are supported:

* `description` - (Optional) The description of the constraint.
* `parameters` - (Required) The constraint parameters, in JSON format. The syntax depends on the constraint type. Refer to the [documentation](https://docs.aws.amazon.com/servicecatalog/latest/dg/API_CreateConstraint.html#API_CreateConstraint_RequestSyntax).
* `portfolio_id` - (Required) The portfolio identifier.
* `product_id` - (Required) The product identifier.
* `type` - (Required) The type of constraint. Valid values: `LAUNCH`, `NOTIFICATION`, `RESOURCE_UPDATE`, `STACKSET` or `TEMPLATE`.

## Attributes Reference

In addition to the arguments, the following attributes are exported:

* `owner` - The owner of the constraint.
* `status` - The status of the current request. Valid values: `AVAILABLE`, `CREATING` or `FAILED`.

## Import

Service Catalog constraints can be imported using their `constraint_id`, e.g.

```bash
$ terraform import aws_servicecatalog_constraint.imported cons-ae6xqmxl4lgfg
```
