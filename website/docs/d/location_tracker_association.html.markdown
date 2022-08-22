---
subcategory: "Location"
layout: "aws"
page_title: "AWS: aws_location_tracker_association"
description: |-
    Retrieve information about a Location Service Tracker Association.
---

# Data Source: aws_location_tracker_association

Retrieve information about a Location Service Tracker Association.

## Example Usage

### Basic Usage

```terraform
data "aws_location_tracker_association" "example" {
  consumer_arn = "arn:aws:geo:region:account-id:geofence-collection/ExampleGeofenceCollectionConsumer"
  tracker_name = "example"
}
```

## Argument Reference

The following arguments are required:

* `consumer_arn` - (Required) The Amazon Resource Name (ARN) of the geofence collection associated to tracker resource.
* `tracker_name` - (Required) The name of the tracker resource associated with a geofence collection.

## Attributes Reference

No additional attributes are exported.
