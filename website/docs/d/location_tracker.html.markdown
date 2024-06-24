---
subcategory: "Location"
layout: "aws"
page_title: "AWS: aws_location_tracker"
description: |-
    Retrieve information about a Location Service Tracker.
---

# Data Source: aws_location_tracker

Retrieve information about a Location Service Tracker.

## Example Usage

```terraform
data "aws_location_tracker" "example" {
  tracker_name = "example"
}
```

## Argument Reference

* `tracker_name` - (Required) Name of the tracker resource.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `create_time` - Timestamp for when the tracker resource was created in ISO 8601 format.
* `description` - Optional description for the tracker resource.
* `kms_key_id` - Key identifier for an AWS KMS customer managed key assigned to the Amazon Location resource.
* `position_filtering` - Position filtering method of the tracker resource.
* `tags` - Key-value map of resource tags for the tracker.
* `tracker_arn` - ARN for the tracker resource. Used when you need to specify a resource across all AWS.
* `update_time` - Timestamp for when the tracker resource was last updated in ISO 8601 format.
