---
subcategory: "Location"
layout: "aws"
page_title: "AWS: aws_location_route_calculator"
description: |-
    Provides a Location Service Route Calculator.
---

# Resource: aws_location_route_calculator

Provides a Location Service Route Calculator.

## Example Usage

```terraform
resource "aws_location_route_calculator" "example" {
  calculator_name = "example"
  data_source     = "Here"
}
```

## Argument Reference

The following arguments are required:

* `calculator_name` - (Required) The name of the route calculator resource.
* `data_source` - (Required) Specifies the data provider of traffic and road network data.

The following arguments are optional:

* `description` - (Optional) The optional description for the route calculator resource.
* `tags` - (Optional) Key-value tags for the route calculator. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `calculator_arn` - The Amazon Resource Name (ARN) for the Route calculator resource. Use the ARN when you specify a resource across AWS.
* `create_time` - The timestamp for when the route calculator resource was created in ISO 8601 format.
* `update_time` - The timestamp for when the route calculator resource was last update in ISO 8601.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `30m`)
* `update` - (Default `30m`)
* `delete` - (Default `30m`)

## Import

`aws_location_route_calculator` can be imported using the route calculator name, e.g.,

```
$ terraform import aws_location_route_calculator.example example
```
