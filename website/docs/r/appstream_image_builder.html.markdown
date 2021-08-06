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

```hcl
resource "aws_appstream_image_builder" "test_fleet" {
  name       = "test-image builder"
  access_endpoints {
    endpoint_type = "STREAMING"
  }
  description                    = "test image builder"
  display_name                   = "test-image builder"
  enable_default_internet_access = false
  image_name                     = "Amazon-AppStream2-Sample-Image-02-04-2019"
  instance_type                  = "stream.standard.large"
  vpc_config {
    subnet_ids                     = ["subnet-06e9b13400c225127"]
  }
  tags = {
    TagName = "tag-value"
  }
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) A unique name for the image builder.
* `name_prefix` -  (Optional) Creates a unique name beginning with the specified prefix. Conflicts with `name`.
* `access_endpoints` - (Optional) The list of interface VPC endpoint (interface endpoint) objects.
  * `endpoint_type` - (Required) The type of interface endpoint.
  * `vpce_id` - (Optional) The identifier (ID) of the VPC in which the interface endpoint is used.
* `appstream_agent_version` - (Optional) The version of the AppStream 2.0 agent to use for this image builder.
* `description` - (Optional) The description to display.
* `display_name` - (Optional) Human-readable friendly name for the AppStream image builder.
* `domain_join_info` - (Optional) The name of the directory and organizational unit (OU) to use to join the image builder to a Microsoft Active Directory domain.
  * `directory_name` - (Optional) The fully qualified name of the directory (for example, corp.example.com).
  * `organizational_unit_distinguished_name` - (Optional) The distinguished name of the organizational unit for computer accounts.
* `enable_default_internet_access` - (Optional) Enables or disables default internet access for the image builder.
* `iam_role_arn` - (Optional) The Amazon Resource Name (ARN) of the IAM role to apply to the image builder.
* `image_name` - (Optional) The name of the image used to create the image builder.
* `image_arn` - (Optional) The ARN of the public, private, or shared image to use.
* `instance_type` - (Required) The instance type to use when launching the image builder.
* `state` - (Optional) The state of the image builder. Valid values are `RUNNING`, `STOPPED`.
* `vpc_config` - (Optional) The VPC configuration for the image builder.
  * `security_group_ids` - The identifiers of the security groups for the image builder or image builder.
  * `subnet_ids` - The identifiers of the subnets to which a network interface is attached from the image builder instance or image builder instance.
* `tags` - Map of tags to attach to AppStream instances.

## Attributes Reference

* `id` - The unique identifier (ID) of the appstream image builder.
* `arn` - The Amazon Resource Name (ARN) of the appstream image builder.
