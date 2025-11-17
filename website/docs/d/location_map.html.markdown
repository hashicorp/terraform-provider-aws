---
subcategory: "Location"
layout: "aws"
page_title: "AWS: aws_location_map"
description: |-
    Retrieve information about a Location Service Map.
---

# Data Source: aws_location_map

Retrieve information about a Location Service Map.

## Example Usage

```terraform
data "aws_location_map" "example" {
  map_name = "example"
}
```

## Argument Reference

This data source supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `map_name` - (Required) Name of the map resource.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `configuration` - List of configurations that specify the map tile style selected from a partner data provider.
    * `style` - The map style selected from an available data provider.
* `create_time` - Timestamp for when the map resource was created in ISO 8601 format.
* `description` - Optional description for the map resource.
* `map_arn` - ARN for the map resource.
* `tags` - Key-value map of resource tags for the map.
* `update_time` - Timestamp for when the map resource was last updated in ISO 8601 format.
