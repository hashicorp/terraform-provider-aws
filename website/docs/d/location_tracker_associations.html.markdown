---
subcategory: "Location"
layout: "aws"
page_title: "AWS: aws_location_tracker_associations"
description: |-
    Retrieve information about Location Service Tracker Associations.
---

# Data Source: aws_location_tracker_associations

Retrieve information about Location Service Tracker Associations.

## Example Usage

### Basic Usage

```terraform
data "aws_location_tracker_associations" "example" {
  tracker_name = "example"
}
```

## Argument Reference

The following arguments are required:

* `tracker_name` - (Required) Name of the tracker resource associated with a geofence collection.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `consumer_arns` - List of geofence collection ARNs associated to the tracker resource.
