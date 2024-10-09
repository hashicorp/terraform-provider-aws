---
subcategory: "Location"
layout: "aws"
page_title: "AWS: aws_location_geofence_collection"
description: |-
    Retrieve information about a Location Service Geofence Collection.
---

# Data Source: aws_location_geofence_collection

Retrieve information about a Location Service Geofence Collection.

## Example Usage

### Basic Usage

```terraform
data "aws_location_geofence_collection" "example" {
  collection_name = "example"
}
```

## Argument Reference

The following arguments are required:

* `collection_name` - (Required) Name of the geofence collection.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `collection_arn` - ARN for the geofence collection resource. Used when you need to specify a resource across all AWS.
* `create_time` - Timestamp for when the geofence collection resource was created in ISO 8601 format.
* `description` - Optional description of the geofence collection resource.
* `kms_key_id` - Key identifier for an AWS KMS customer managed key assigned to the Amazon Location resource.
* `tags` - Key-value map of resource tags for the geofence collection.
* `update_time` - Timestamp for when the geofence collection resource was last updated in ISO 8601 format.
