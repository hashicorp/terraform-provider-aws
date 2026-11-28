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

This data source supports the following arguments:

* `account` - (Optional) AWS account ID where this data source will be read. Defaults to the account of the provider's credentials. See the [Per-Resource Account guide](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/guides/per-resource-account).
* `consumer_arn` - (Required) ARN of the geofence collection associated to tracker resource.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `tracker_name` - (Required) Name of the tracker resource associated with a geofence collection.

## Attribute Reference

This data source exports no additional attributes.
