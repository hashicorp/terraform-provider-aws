---
subcategory: "Location"
layout: "aws"
page_title: "AWS: aws_location_place_index"
description: |-
    Retrieve information about a Location Service Place Index.
---

# Data Source: aws_location_place_index

Retrieve information about a Location Service Place Index.

## Example Usage

```terraform
data "aws_location_place_index" "example" {
  index_name = "example"
}
```

## Argument Reference

* `index_name` - (Required) The name of the place index resource.

## Attribute Reference

* `create_time` - The timestamp for when the place index resource was created in ISO 8601 format.
* `data_source` - The data provider of geospatial data.
* `data_source_configuration` - List of configurations that specify data storage option for requesting Places.
* `description` - The optional description for the place index resource.
* `index_arn` - The Amazon Resource Name (ARN) for the place index resource.
* `tags` - Key-value map of resource tags for the map.
* `update_time` - The timestamp for when the place index resource was last update in ISO 8601.
