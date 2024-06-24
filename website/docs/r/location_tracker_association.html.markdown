---
subcategory: "Location"
layout: "aws"
page_title: "AWS: aws_location_tracker_association"
description: |-
  Terraform resource for managing an AWS Location Tracker Association.
---

# Resource: aws_location_tracker_association

Terraform resource for managing an AWS Location Tracker Association.

## Example Usage

```terraform
resource "aws_location_geofence_collection" "example" {
  collection_name = "example"
}

resource "aws_location_tracker" "example" {
  tracker_name = "example"
}

resource "aws_location_tracker_association" "example" {
  consumer_arn = aws_location_geofence_collection.example.collection_arn
  tracker_name = aws_location_tracker.example.tracker_name
}
```

## Argument Reference

The following arguments are required:

* `consumer_arn` - (Required) The Amazon Resource Name (ARN) for the geofence collection to be associated to tracker resource. Used when you need to specify a resource across all AWS.
* `tracker_name` - (Required) The name of the tracker resource to be associated with a geofence collection.

## Attribute Reference

This resource exports no additional attributes.

## Timeouts

`aws_location_tracker_association` provides the following [Timeouts](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts) configuration options:

* `create` - (Optional, Default: `30m`)
* `delete` - (Optional, Default: `30m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Location Tracker Association using the `tracker_name|consumer_arn`. For example:

```terraform
import {
  to = aws_location_tracker_association.example
  id = "tracker_name|consumer_arn"
}
```

Using `terraform import`, import Location Tracker Association using the `tracker_name|consumer_arn`. For example:

```console
% terraform import aws_location_tracker_association.example "tracker_name|consumer_arn"
```
