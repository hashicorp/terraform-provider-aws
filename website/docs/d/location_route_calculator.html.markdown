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

* `calculator_name` - (Required) The name of the route calculator resource.

## Attributes Reference

* `calculator_arn` - The Amazon Resource Name (ARN) for the Route calculator resource. Use the ARN when you specify a resource across AWS.
* `create_time` - The timestamp for when the route calculator resource was created in ISO 8601 format.
* `data_source` - The data provider of traffic and road network data.
* `description` - The optional description of the route calculator resource.
* `tags` - Key-value map of resource tags for the route calculator.
* `update_time` - The timestamp for when the route calculator resource was last updated in ISO 8601 format.
