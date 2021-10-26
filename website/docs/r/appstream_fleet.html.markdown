---
subcategory: "AppStream"
layout: "aws"
page_title: "AWS: aws_appstream_fleet"
description: |-
  Provides an AppStream fleet
---

# Resource: aws_appstream_fleet

Provides an AppStream fleet.

## Example Usage

```terraform
resource "aws_appstream_fleet" "test_fleet" {
  name = "test-fleet"

  compute_capacity {
    desired_instances = 1
  }

  description                        = "test fleet"
  idle_disconnect_timeout_in_seconds = 15
  display_name                       = "test-fleet"
  enable_default_internet_access     = false
  fleet_type                         = "ON_DEMAND"
  image_name                         = "Amazon-AppStream2-Sample-Image-02-04-2019"
  instance_type                      = "stream.standard.large"
  max_user_duration_in_seconds       = 600

  vpc_config {
    subnet_ids = ["subnet-06e9b13400c225127"]
  }

  tags = {
    TagName = "tag-value"
  }
}
```

## Argument Reference

The following arguments are required:

* `compute_capacity` - (Required) Configuration block for the desired capacity of the fleet. See below.
* `instance_type` - (Required) Instance type to use when launching fleet instances.
* `name` - (Required) Unique name for the fleet.

The following arguments are optional:

* `description` - (Optional) Description to display.
* `disconnect_timeout_in_seconds` - (Optional) Amount of time that a streaming session remains active after users disconnect.
* `display_name` - (Optional) Human-readable friendly name for the AppStream fleet.
* `domain_join_info` - (Optional) Configuration block for the name of the directory and organizational unit (OU) to use to join the fleet to a Microsoft Active Directory domain. See below.
* `enable_default_internet_access` - (Optional) Enables or disables default internet access for the fleet.
* `fleet_type` - (Optional) Fleet type. Valid values are: `ON_DEMAND`, `ALWAYS_ON`
* `iam_role_arn` - (Optional) ARN of the IAM role to apply to the fleet.
* `idle_disconnect_timeout_in_seconds` - (Optional) Amount of time that users can be idle (inactive) before they are disconnected from their streaming session and the `disconnect_timeout_in_seconds` time interval begins.
* `image_name` - (Optional) Name of the image used to create the fleet.
* `image_arn` - (Optional) ARN of the public, private, or shared image to use.
* `stream_view` - (Optional) AppStream 2.0 view that is displayed to your users when they stream from the fleet. When `APP` is specified, only the windows of applications opened by users display. When `DESKTOP` is specified, the standard desktop that is provided by the operating system displays.
* `max_user_duration_in_seconds` - (Optional) Maximum amount of time that a streaming session can remain active, in seconds.
* `vpc_config` - (Optional) Configuration block for the VPC configuration for the image builder. See below.
* `tags` - (Optional) Map of tags to attach to AppStream instances.

### `compute_capacity`

* `desired_instances` - (Required) Desired number of streaming instances.

### `domain_join_info`

* `directory_name` - (Optional) Fully qualified name of the directory (for example, corp.example.com).
* `organizational_unit_distinguished_name` - (Optional) Distinguished name of the organizational unit for computer accounts.

### `vpc_config`

* `security_group_ids` - Identifiers of the security groups for the fleet or image builder.
* `subnet_ids` - Identifiers of the subnets to which a network interface is attached from the fleet instance or image builder instance.


## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - Unique identifier (ID) of the appstream fleet.
* `arn` - ARN of the appstream fleet.
* `state` - State of the fleet. Can be `STARTING`, `RUNNING`, `STOPPING` or `STOPPED`
* `created_time` -  Date and time, in UTC and extended RFC 3339 format, when the fleet was created.
* `compute_capacity` - Describes the capacity status for a fleet.

### `compute_capacity`

* `available` - Number of currently available instances that can be used to stream sessions.
* `in_use` - Number of instances in use for streaming.
* `running` - Total number of simultaneous streaming instances that are running.


## Import

`aws_appstream_fleet` can be imported using the id, e.g.,

```
$ terraform import aws_appstream_fleet.example fleetNameExample
```
