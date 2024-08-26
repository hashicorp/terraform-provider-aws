---
subcategory: "Auto Scaling"
layout: "aws"
page_title: "AWS: aws_launch_configuration"
description: |-
  Provides a Launch Configuration data source.
---

# Data Source: aws_launch_configuration

Provides information about a Launch Configuration.

## Example Usage

```terraform
data "aws_launch_configuration" "ubuntu" {
  name = "test-launch-config"
}
```

## Argument Reference

This data source supports the following arguments:

* `name` - (Required) Name of the launch configuration.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `id` - ID of the launch configuration.
* `arn` - Amazon Resource Name of the launch configuration.
* `name` - Name of the launch configuration.
* `image_id` - EC2 Image ID of the instance.
* `instance_type` - Instance Type of the instance to launch.
* `iam_instance_profile` - The IAM Instance Profile to associate with launched instances.
* `key_name` - Key Name that should be used for the instance.
* `metadata_options` - Metadata options for the instance.
    * `http_endpoint` - State of the metadata service: `enabled`, `disabled`.
    * `http_tokens` - If session tokens are required: `optional`, `required`.
    * `http_put_response_hop_limit` - The desired HTTP PUT response hop limit for instance metadata requests.
* `security_groups` - List of associated Security Group IDS.
* `associate_public_ip_address` - Whether a Public IP address is associated with the instance.
* `primary_ipv6` - Whether the first IPv6 GUA will be made the primary IPv6 address.
* `user_data` - User Data of the instance.
* `enable_monitoring` - Whether Detailed Monitoring is Enabled.
* `ebs_optimized` - Whether the launched EC2 instance will be EBS-optimized.
* `root_block_device` - Root Block Device of the instance.
* `ebs_block_device` - EBS Block Devices attached to the instance.
* `ephemeral_block_device` - The Ephemeral volumes on the instance.
* `spot_price` - Price to use for reserving Spot instances.
* `placement_tenancy` - Tenancy of the instance.

`root_block_device` is exported with the following attributes:

* `delete_on_termination` - Whether the EBS Volume will be deleted on instance termination.
* `encrypted` - Whether the volume is Encrypted.
* `iops` - Provisioned IOPs of the volume.
* `throughput` - Throughput of the volume.
* `volume_size` - Size of the volume.
* `volume_type` - Type of the volume.

`ebs_block_device` is exported with the following attributes:

* `delete_on_termination` - Whether the EBS Volume will be deleted on instance termination.
* `device_name` - Name of the device.
* `encrypted` - Whether the volume is Encrypted.
* `iops` - Provisioned IOPs of the volume.
* `no_device` - Whether the device in the block device mapping of the AMI is suppressed.
* `snapshot_id` - Snapshot ID of the mount.
* `throughput` - Throughput of the volume.
* `volume_size` - Size of the volume.
* `volume_type` - Type of the volume.

`ephemeral_block_device` is exported with the following attributes:

* `device_name` - Name of the device.
* `virtual_name` - Virtual Name of the device.
