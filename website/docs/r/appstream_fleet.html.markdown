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

```hcl
resource "aws_appstream_fleet" "test_fleet" {
  name       = "test-fleet"
  compute_capacity {
    desired_instances = 1
  }
  description                    = "test fleet"
  idle_disconnect_timeout_in_seconds             = 15
  display_name                   = "test-fleet"
  enable_default_internet_access = false
  fleet_type                     = "ON_DEMAND"
  image_name                     = "Amazon-AppStream2-Sample-Image-02-04-2019"
  instance_type                  = "stream.standard.large"
  max_user_duration_in_seconds              = 600
  vpc_config {
    subnet_ids                     = ["subnet-06e9b13400c225127"]
    security_group_ids             = ["sg-0397cdfe509785903", "sg-0bd2dddff01dee52d"]
  }
  tags = {
    TagName = "tag-value"
  }
  state = "RUNNING"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) A unique name for the fleet.
* `name_prefix` -  (Optional) Creates a unique name beginning with the specified prefix. Conflicts with `name`.
* `compute_capacity` - (Required) The desired capacity for the fleet.
  * `desired_instances` - (Required) The desired number of streaming instances.
* `description` - (Optional) The description to display.
* `disconnect_timeout_in_seconds` - (Optional) The amount of time that a streaming session remains active after users disconnect.
* `display_name` - (Optional) Human-readable friendly name for the AppStream fleet.
* `domain_join_info` - (Optional) The name of the directory and organizational unit (OU) to use to join the fleet to a Microsoft Active Directory domain.
  * `directory_name` - (Optional) The fully qualified name of the directory (for example, corp.example.com).
  * `organizational_unit_distinguished_name` - (Optional) The distinguished name of the organizational unit for computer accounts.
* `enable_default_internet_access` - (Optional) Enables or disables default internet access for the fleet.
* `fleet_type` - (Optional) The fleet type. Valid values are: `ON_DEMAND`, `ALWAYS_ON`
* `iam_role_arn` - (Optional) The Amazon Resource Name (ARN) of the IAM role to apply to the fleet.
* `idle_disconnect_timeout_in_seconds` - (Optional) The amount of time that users can be idle (inactive) before they are disconnected from their streaming session and the `disconnect_timeout_in_seconds` time interval begins.
* `image_name` - (Optional) The name of the image used to create the fleet.
* `image_arn` - (Optional) The ARN of the public, private, or shared image to use.
* `instance_type` - (Required) The instance type to use when launching fleet instances.
* `stream_view` - (Optional) The AppStream 2.0 view that is displayed to your users when they stream from the fleet. When `APP` is specified, only the windows of applications opened by users display. When `DESKTOP` is specified, the standard desktop that is provided by the operating system displays.
* `max_user_duration_in_seconds` - (Optional) The maximum amount of time that a streaming session can remain active, in seconds.
* `stack_name` - (Optional) Name of AppStream stack, to be associated with this fleet.
* `state` - (Optional) The state of the fleet. Valid values are `RUNNING`, `STOPPED`.
* `vpc_config` - (Optional) The VPC configuration for the image builder.
  * `security_group_ids` - The identifiers of the security groups for the fleet or image builder.
  * `subnet_ids` - The identifiers of the subnets to which a network interface is attached from the fleet instance or image builder instance.
* `tags` - Map of tags to attach to AppStream instances.

## Attributes Reference

* `id` - The unique identifier (ID) of the appstream fleet.
* `arn` - The Amazon Resource Name (ARN) of the appstream fleet.
