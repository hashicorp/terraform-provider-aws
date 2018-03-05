---
layout: "aws"
page_title: "AWS: aws_launch_configuration"
sidebar_current: "docs-aws-datasource-launch-configuration"
description: |-
  Provides a Launch Configuration data source.
---

# aws_launch_configuration

Provides information about a Launch Configuration.

## Example Usage

```hcl
data "aws_launch_configuration" "ubuntu" {
  name = "test-launch-config"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the launch configuration.

## Attributes Reference

The following attributes are exported:

* `id` - The ID of the launch configuration.
* `name` - The name of the launch configuration.
* `image_id` - The EC2 image ID of the instance.
* `instance_type` - The type of the instance to launch.
* `iam_instance_profile` - The IAM instance profile to associate with launched instances.
* `key_name` - The key name that should be used for the instance.
* `security_groups` - A list of associated security group IDS.
* `associate_public_ip_address` - Whether a public ip address is associated with the instance.
* `vpc_classic_link_id` - The ID of a ClassicLink-enabled VPC.
* `vpc_classic_link_security_groups` - The IDs of one or more security groups for the specified ClassicLink-enabled VPC.
* `user_data` - The user data of the instance.
* `enable_monitoring` - Whether detailed monitoring is enabled.
* `ebs_optimized` - Whether the launched EC2 instance will be EBS-optimized.
* `root_block_device` - The root block device of the instance.
* `ebs_block_device` - The EBS block devices attached to the instance.
* `ephemeral_block_device` - The Ephemeral volumes on the instance.
* `spot_price` - The price to use for reserving spot instances.
* `placement_tenancy` - The tenancy of the instance.
