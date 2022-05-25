---
subcategory: "Location"
layout: "aws"
page_title: "AWS: aws_location_map"
description: |-
    Provides a Location Service Map.
---

# Resource: aws_location_map

Provides a Location Service Map.

## Example Usage

```terraform
resource "aws_location_map" "example" {
  configuration {
    style = "VectorHereBerlin"
  }

  map_name = "example"
}
```

## Argument Reference

The following arguments are required:

* `configuration` - (Required) Configuration block with the map style selected from an available data provider. Detailed below.
* `map_name` - (Required) The name for the map resource.

The following arguments are optional:

* `description` - (Optional) An optional description for the map resource.
* `tags` - (Optional) Key-value tags for the map. If configured with a provider [`default_tags` configuration block](https://www.terraform.io/docs/providers/aws/index.html#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

### configuration

The following arguments are required:

* `style` - (Required) Specifies the map style selected from an available data provider. Valid values can be found in the [Location Service CreateMap API Reference](https://docs.aws.amazon.com/location-maps/latest/APIReference/API_CreateMap.html).

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `create_time` - The timestamp for when the map resource was created in ISO 8601 format.
* `map_arn` - The Amazon Resource Name (ARN) for the map resource. Used to specify a resource across all AWS.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://www.terraform.io/docs/providers/aws/index.html#default_tags-configuration-block).
* `update_time` - The timestamp for when the map resource was last updated in ISO 8601.

## Import

`aws_location_map` resources can be imported using the map name, e.g.:

```
$ terraform import aws_location_map.example example
```
