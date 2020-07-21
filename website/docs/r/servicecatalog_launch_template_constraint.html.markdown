---
subcategory: "Service Catalog"
layout: "aws"
page_title: "AWS: aws_servicecatalog_launch_template_constraint"
description: |-
  Provides a resource to control the template constraints which restrict launching a product in a specific portfolio
---

# Resource: aws_servicecatalog_launch_template_constraint

Provides a resource to control the template constraints which restrict launching a product in a specific portfolio.

## Example Usage

```hcl
resource "aws_servicecatalog_launch_template_constraint" "sample_constraint" {
  description = "Sample template constraint rules"
  rule {
    name = "rule01"
    rule_condition = jsonencode({
     "Fn::Equals" = [
       {"Ref" = "Environment"},
       "test"
      ]
    })
    assertion {
     assert = jsonencode({
       "Fn::Contains" = [
         ["m1.small"],
         {"Ref" = "InstanceType"}
       ]
     })
     assert_description = "For the test environment, the instance type must be m1.small"
    }
  }
  rule {
   name = "rule02"
   rule_condition = jsonencode({
     "Fn::Equals" = [
       {"Ref" = "Environment"}, 
       "prod"
     ]
   })
   assertion {
     assert = jsonencode({
       "Fn::Contains" = [
         ["m1.large"],
         {"Ref" = "InstanceType"} 
       ]
     })
     assert_description = "For the prod environment, the instance type must be m1.large"
   }
  }
  portfolio_id = aws_servicecatalog_portfolio.myportfolio.id
  product_id = aws_servicecatalog_product.myproduct.id
}
```

## Argument Reference

The following arguments are supported:

* `description` - (Optional) The description of the constraint.
* `rule` - (Required) The rules to define the template constraint.
* `portfolio_id` - (Required) The portfolio identifier.
* `product_id` - (Required) The product identifier.

## Attributes Reference

In addition to the arguments, the following attributes are exported:

* `owner` - The owner of the constraint.
* `status` - The status of the current request. Valid values: `AVAILABLE`, `CREATING` or `FAILED`.
* `parameters` - The constraint parameters, in JSON format. The syntax depends on the constraint type. Refer to the [documentation](https://docs.aws.amazon.com/servicecatalog/latest/dg/API_CreateConstraint.html#API_CreateConstraint_RequestSyntax).
* `type` - The type of constraint. Valid values: `LAUNCH`

## Import

Service Catalog launch template constraints can be imported using their `constraint_id`, e.g.

```bash
$ terraform import aws_servicecatalog_launch_template_constraint.imported cons-ae6xqmxl4lgfg
```
