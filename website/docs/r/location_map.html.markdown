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

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `description` - (Optional) An optional description for the map resource.
* `tags` - (Optional) Key-value tags for the map. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

### configuration

The following arguments are required:

* `style` - (Required) Specifies the map style selected from an available data provider. Valid values can be found in the [Location Service CreateMap API Reference](https://docs.aws.amazon.com/location/latest/APIReference/API_CreateMap.html).

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `create_time` - The timestamp for when the map resource was created in ISO 8601 format.
* `map_arn` - The Amazon Resource Name (ARN) for the map resource. Used to specify a resource across all AWS.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).
* `update_time` - The timestamp for when the map resource was last updated in ISO 8601 format.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import `aws_location_map` resources using the map name. For example:

```terraform
import {
  to = aws_location_map.example
  id = "example"
}
```

Using `terraform import`, import `aws_location_map` resources using the map name. For example:

```console
% terraform import aws_location_map.example example
```
