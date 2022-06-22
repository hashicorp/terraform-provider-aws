---
subcategory: "Location"
layout: "aws"
page_title: "AWS: aws_location_tracker"
description: |-
    Provides a Location Service Tracker.
---

# Resource: aws_location_tracker

Provides a Location Service Tracker.

## Example Usage

```terraform
resource "aws_location_tracker" "example" {
  tracker_name = "example"
}
```

## Argument Reference

The following arguments are required:

* `tracker_name` - (Required) The name of the tracker resource.

The following arguments are optional:

* `description` - (Optional) The optional description for the tracker resource.
* `kms_key_id` - (Optional) A key identifier for an AWS KMS customer managed key assigned to the Amazon Location resource.
* `position_filtering` - (Optional) The position filtering method of the tracker resource. Valid values: `TimeBased`, `DistanceBased`, `AccuracyBased`. Default: `TimeBased`.
* `tags` - (Optional) Key-value tags for the place index. If configured with a provider [`default_tags` configuration block](https://www.terraform.io/docs/providers/aws/index.html#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `create_time` - The timestamp for when the tracker resource was created in ISO 8601 format.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://www.terraform.io/docs/providers/aws/index.html#default_tags-configuration-block).
* `tracker_arn` - The Amazon Resource Name (ARN) for the tracker resource. Used when you need to specify a resource across all AWS.
* `update_time` - The timestamp for when the tracker resource was last update in ISO 8601 format.

## Import

`aws_location_tracker` resources can be imported using the tracker name, e.g.:

```
$ terraform import aws_location_tracker.example example
```
