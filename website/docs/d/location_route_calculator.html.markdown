---
subcategory: "Location"
layout: "aws"
page_title: "AWS: aws_location_route_calculator"
description: |-
    Retrieve information about a Location Service Route Calculator.
---

# Data Source: aws_location_route_calculator

Retrieve information about a Location Service Route Calculator.

## Example Usage

```terraform
data "aws_location_route_calculator" "example" {
  calculator_name = "example"
}
```

## Argument Reference

This data source supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `calculator_name` - (Required) Name of the route calculator resource.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `calculator_arn` - ARN for the Route calculator resource. Use the ARN when you specify a resource across AWS.
* `create_time` - Timestamp for when the route calculator resource was created in ISO 8601 format.
* `data_source` - Data provider of traffic and road network data.
* `description` - Optional description of the route calculator resource.
* `tags` - Key-value map of resource tags for the route calculator.
* `update_time` - Timestamp for when the route calculator resource was last updated in ISO 8601 format.
