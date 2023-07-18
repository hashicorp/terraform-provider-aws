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

* `index_name` - (Required) Name of the place index resource.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `create_time` - Timestamp for when the place index resource was created in ISO 8601 format.
* `data_source` - Data provider of geospatial data.
* `data_source_configuration` - List of configurations that specify data storage option for requesting Places.
* `description` - Optional description for the place index resource.
* `index_arn` - ARN for the place index resource.
* `tags` - Key-value map of resource tags for the place index.
* `update_time` - Timestamp for when the place index resource was last updated in ISO 8601 format.
