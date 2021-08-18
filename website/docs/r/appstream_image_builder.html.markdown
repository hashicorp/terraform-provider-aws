---
subcategory: "AppStream"
layout: "aws"
page_title: "AWS: aws_appstream_image_builder"
description: |-
  Provides an AppStream image builder
---

# Resource: aws_appstream_image_builder

Provides an AppStream image builder.

## Example Usage

```terraform
resource "aws_appstream_image_builder" "test_fleet" {
  name                           = "Image Builder Name"
  description                    = "Description of a ImageBuilder"
  display_name                   = "Display name of a ImageBuilder"
  enable_default_internet_access = false
  image_name                     = "Amazon-AppStream2-Sample-Image-02-04-2019"
  instance_type                  = "stream.standard.large"

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

* `instance_type` - (Required) The instance type to use when launching the image builder.
* `name` - (Required) Unique name for the image builder.

The following arguments are optional:

* `access_endpoints` - (Optional) Configuration block for the list of interface VPC endpoint (interface endpoint) objects. See below.
* `appstream_agent_version` - (Optional) The version of the AppStream 2.0 agent to use for this image builder.
* `description` - (Optional) Description to display.
* `display_name` - (Optional) Human-readable friendly name for the AppStream image builder.
* `domain_join_info` - (Optional) Configuration block for the name of the directory and organizational unit (OU) to use to join the image builder to a Microsoft Active Directory domain. See below.
* `enable_default_internet_access` - (Optional) Enables or disables default internet access for the image builder.
* `iam_role_arn` - (Optional) ARN of the IAM role to apply to the image builder.
* `image_name` - (Optional) Name of the image used to create the image builder.
* `image_arn` - (Optional) ARN of the public, private, or shared image to use.
* `vpc_config` - (Optional) Configuration block for the VPC configuration for the image builder. See below.
* `tags` - (Optional) Map of tags to attach to AppStream instances.


### `access_endpoints`

* `endpoint_type` - (Required) The type of interface endpoint.
* `vpce_id` - (Optional) The identifier (ID) of the VPC in which the interface endpoint is used.

### `domain_join_info`

* `directory_name` - (Optional) The fully qualified name of the directory (for example, corp.example.com).
* `organizational_unit_distinguished_name` - (Optional) The distinguished name of the organizational unit for computer accounts.

### `vpc_config`

* `security_group_ids` - (Optional) The identifiers of the security groups for the image builder or image builder.
* `subnet_ids` - (Optional) The identifiers of the subnets to which a network interface is attached from the image builder instance or image builder instance.


## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - Unique identifier (ID) of the appstream image builder.
* `arn` - ARN of the appstream image builder.
* `state` - State of the image builder. Can be: `PENDING`, `UPDATING_AGENT`, `RUNNING`, `STOPPING`, `STOPPED`, `REBOOTING`, `SNAPSHOTTING`, `DELETING`, `FAILED`, `UPDATING`, `PENDING_QUALIFICATION`
* `created_time` -  Date and time, in UTC and extended RFC 3339 format, when the image builder was created.


## Import

`aws_appstream_image_builder` can be imported using the id, e.g.

```
$ terraform import aws_appstream_image_builder.example imageBuilderExample