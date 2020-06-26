---
subcategory: "Service Catalog"
layout: "aws"
page_title: "AWS: aws_servicecatalog_launch_role_constraint"
description: |-
  Provides a resource to control the roles that are permitted to launch a product in a specific portfolio
---

# Resource: aws_servicecatalog_launch_role_constraint

Provides a resource to control the roles that are permitted to launch a product in a specific portfolio.

If you specify the `local_role_name` property, when an account uses the launch constraint, the IAM role with that name in the account will be used. This allows launch-role constraints to be account-agnostic so the administrator can create fewer resources per shared account. 

## Example Usage

```hcl
resource "aws_servicecatalog_launch_role_constraint" "launch_role_constraint" {
  description = "Only Team Alpha Admins may launch"
  local_role_name=  "teams/alpha/Admin"
  portfolio_id = aws_servicecatalog_portfolio.myportfolio.id
  product_id = aws_servicecatalog_product.myproduct.id
}
```

## Argument Reference

The following arguments are supported:

* `description` - (Optional) The description of the constraint.
* `local_role_name` - (Optional) The local name of the role.
* `role_arn` - (Optional) The ARN of the role.
* `portfolio_id` - (Required) The portfolio identifier.
* `product_id` - (Required) The product identifier.

You are required to specify either the `role_arn` or the `local_role_name` but can't use both. 

## Attributes Reference

In addition to the arguments, the following attributes are exported:

* `owner` - The owner of the constraint.
* `status` - The status of the current request. Valid values: `AVAILABLE`, `CREATING` or `FAILED`.
* `parameters` - The constraint parameters, in JSON format. The syntax depends on the constraint type. Refer to the [documentation](https://docs.aws.amazon.com/servicecatalog/latest/dg/API_CreateConstraint.html#API_CreateConstraint_RequestSyntax).
* `type` - The type of constraint. Valid values: `LAUNCH`

## Import

Service Catalog launch role constraints can be imported using their `constraint_id`, e.g.

```bash
$ terraform import aws_servicecatalog_launch_role_constraint.imported cons-ae6xqmxl4lgfg
```
