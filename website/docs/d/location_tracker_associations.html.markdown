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

This data source supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `tracker_name` - (Required) Name of the tracker resource associated with a geofence collection.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `consumer_arns` - List of geofence collection ARNs associated to the tracker resource.
