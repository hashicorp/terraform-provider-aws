---
subcategory: "Auto Scaling"
layout: "aws"
page_title: "AWS: aws_launch_configuration"
description: |-
  Provides a resource to create a new launch configuration, used for autoscaling groups.
---

# Resource: aws_launch_configuration

Provides a resource to create a new launch configuration, used for autoscaling groups.

!> **WARNING:** The use of launch configurations is discouraged in favor of launch templates. Read more in the [AWS EC2 Documentation](https://docs.aws.amazon.com/autoscaling/ec2/userguide/launch-configurations.html).

-> **Note** When using `aws_launch_configuration` with `aws_autoscaling_group`, it is recommended to use the `name_prefix` (Optional) instead of the `name` (Optional) attribute. This will allow Terraform lifecycles to detect changes to the launch configuration and update the autoscaling group correctly.

## Example Usage

```terraform
data "aws_ami" "ubuntu" {
  most_recent = true

  filter {
    name   = "name"
    values = ["ubuntu/images/hvm-ssd/ubuntu-trusty-14.04-amd64-server-*"]
  }

  filter {
    name   = "virtualization-type"
    values = ["hvm"]
  }

  owners = ["099720109477"] # Canonical
}

resource "aws_launch_configuration" "as_conf" {
  name          = "web_config"
  image_id      = data.aws_ami.ubuntu.id
  instance_type = "t2.micro"
}
```

## Using with AutoScaling Groups

Launch Configurations cannot be updated after creation with the Amazon
Web Service API. In order to update a Launch Configuration, Terraform will
destroy the existing resource and create a replacement. In order to effectively
use a Launch Configuration resource with an [AutoScaling Group resource][1],
it's recommended to specify `create_before_destroy` in a [lifecycle][2] block.
Either omit the Launch Configuration `name` attribute, or specify a partial name
with `name_prefix`.  Example:

```terraform
data "aws_ami" "ubuntu" {
  most_recent = true

  filter {
    name   = "name"
    values = ["ubuntu/images/hvm-ssd/ubuntu-trusty-14.04-amd64-server-*"]
  }

  filter {
    name   = "virtualization-type"
    values = ["hvm"]
  }

  owners = ["099720109477"] # Canonical
}

resource "aws_launch_configuration" "as_conf" {
  name_prefix   = "terraform-lc-example-"
  image_id      = data.aws_ami.ubuntu.id
  instance_type = "t2.micro"

  lifecycle {
    create_before_destroy = true
  }
}

resource "aws_autoscaling_group" "bar" {
  name                 = "terraform-asg-example"
  launch_configuration = aws_launch_configuration.as_conf.name
  min_size             = 1
  max_size             = 2

  lifecycle {
    create_before_destroy = true
  }
}
```

With this setup Terraform generates a unique name for your Launch
Configuration and can then update the AutoScaling Group without conflict before
destroying the previous Launch Configuration.

## Using with Spot Instances

Launch configurations can set the spot instance pricing to be used for the
Auto Scaling Group to reserve instances. Simply specifying the `spot_price`
parameter will set the price on the Launch Configuration which will attempt to
reserve your instances at this price.  See the [AWS Spot Instance
documentation](https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/using-spot-instances.html)
for more information or how to launch [Spot Instances][3] with Terraform.

```terraform
data "aws_ami" "ubuntu" {
  most_recent = true

  filter {
    name   = "name"
    values = ["ubuntu/images/hvm-ssd/ubuntu-trusty-14.04-amd64-server-*"]
  }

  filter {
    name   = "virtualization-type"
    values = ["hvm"]
  }

  owners = ["099720109477"] # Canonical
}

resource "aws_launch_configuration" "as_conf" {
  image_id      = data.aws_ami.ubuntu.id
  instance_type = "m4.large"
  spot_price    = "0.001"

  lifecycle {
    create_before_destroy = true
  }
}

resource "aws_autoscaling_group" "bar" {
  name                 = "terraform-asg-example"
  launch_configuration = aws_launch_configuration.as_conf.name
}
```

## Argument Reference

The following arguments are required:

* `image_id` - (Required) The EC2 image ID to launch.
* `instance_type` - (Required) The size of instance to launch.

The following arguments are optional:

* `associate_public_ip_address` - (Optional) Associate a public ip address with an instance in a VPC.
* `ebs_block_device` - (Optional) Additional EBS block devices to attach to the instance. See [Block Devices](#block-devices) below for details.
* `ebs_optimized` - (Optional) If true, the launched EC2 instance will be EBS-optimized.
* `enable_monitoring` - (Optional) Enables/disables detailed monitoring. This is enabled by default.
* `ephemeral_block_device` - (Optional) Customize Ephemeral (also known as "Instance Store") volumes on the instance. See [Block Devices](#block-devices) below for details.
* `iam_instance_profile` - (Optional) The name attribute of the IAM instance profile to associate with launched instances.
* `key_name` - (Optional) The key name that should be used for the instance.
* `metadata_options` - The metadata options for the instance.
    * `http_endpoint` - The state of the metadata service: `enabled`, `disabled`.
    * `http_tokens` - If session tokens are required: `optional`, `required`.
    * `http_put_response_hop_limit` - The desired HTTP PUT response hop limit for instance metadata requests.
* `name` - (Optional) The name of the launch configuration. If you leave this blank, Terraform will auto-generate a unique name. Conflicts with `name_prefix`.
* `name_prefix` - (Optional) Creates a unique name beginning with the specified prefix. Conflicts with `name`.
* `security_groups` - (Optional) A list of associated security group IDS.
* `placement_tenancy` - (Optional) The tenancy of the instance. Valid values are `default` or `dedicated`, see [AWS's Create Launch Configuration](http://docs.aws.amazon.com/AutoScaling/latest/APIReference/API_CreateLaunchConfiguration.html) for more details.
* `root_block_device` - (Optional) Customize details about the root block device of the instance. See [Block Devices](#block-devices) below for details.
* `spot_price` - (Optional; Default: On-demand price) The maximum price to use for reserving spot instances.
* `user_data` - (Optional) The user data to provide when launching the instance. Do not pass gzip-compressed data via this argument; see `user_data_base64` instead.
* `user_data_base64` - (Optional) Can be used instead of `user_data` to pass base64-encoded binary data directly. Use this instead of `user_data` whenever the value is not a valid UTF-8 string. For example, gzip-encoded user data must be base64-encoded and passed via this argument to avoid corruption.

## Block devices

Each of the `*_block_device` attributes controls a portion of the AWS
Launch Configuration's "Block Device Mapping". It's a good idea to familiarize yourself with [AWS's Block Device
Mapping docs](https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/block-device-mapping-concepts.html)
to understand the implications of using these attributes.

Each AWS Instance type has a different set of Instance Store block devices
available for attachment. AWS [publishes a
list](https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/InstanceStorage.html#StorageOnInstanceTypes)
of which ephemeral devices are available on each type. The devices are always
identified by the `virtual_name` in the format `ephemeral{0..N}`.

~> **NOTE:** Changes to `*_block_device` configuration of _existing_ resources
cannot currently be detected by Terraform. After updating to block device
configuration, resource recreation can be manually triggered by using the
[`taint` command](https://www.terraform.io/docs/commands/taint.html).
  
### ebs_block_device

Modifying any of the `ebs_block_device` settings requires resource replacement.

* `device_name` - (Required) The name of the device to mount.
* `snapshot_id` - (Optional) The Snapshot ID to mount.
* `volume_type` - (Optional) The type of volume. Can be `standard`, `gp2`, `gp3`, `st1`, `sc1` or `io1`.
* `volume_size` - (Optional) The size of the volume in gigabytes.
* `iops` - (Optional) The amount of provisioned
  [IOPS](https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/ebs-io-characteristics.html).
  This must be set with a `volume_type` of `"io1"`.
* `throughput` - (Optional) The throughput (MiBps) to provision for a `gp3` volume.
* `delete_on_termination` - (Optional) Whether the volume should be destroyed
  on instance termination (Default: `true`).
* `encrypted` - (Optional) Whether the volume should be encrypted or not. Defaults to `false`.
* `no_device` - (Optional) Whether the device in the block device mapping of the AMI is suppressed.

### ephemeral_block_device

* `device_name` - (Required) The name of the block device to mount on the instance.
* `no_device` - (Optional) Whether the device in the block device mapping of the AMI is suppressed.
* `virtual_name` - (Optional) The [Instance Store Device Name](https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/InstanceStorage.html#InstanceStoreDeviceNames).

### root_block_device

-> Modifying any of the `root_block_device` settings requires resource replacement.

* `delete_on_termination` - (Optional) Whether the volume should be destroyed on instance termination. Defaults to `true`.
* `encrypted` - (Optional) Whether the volume should be encrypted or not. Defaults to `false`.
* `iops` - (Optional) The amount of provisioned [IOPS](https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/ebs-io-characteristics.html). This must be set with a `volume_type` of `io1`.
* `throughput` - (Optional) The throughput (MiBps) to provision for a `gp3` volume.
* `volume_size` - (Optional) The size of the volume in gigabytes.
* `volume_type` - (Optional) The type of volume. Can be `standard`, `gp2`, `gp3`, `st1`, `sc1` or `io1`.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - The ID of the launch configuration.
* `arn` - The Amazon Resource Name of the launch configuration.
* `name` - The name of the launch configuration.

[1]: /docs/providers/aws/r/autoscaling_group.html
[2]: https://www.terraform.io/docs/configuration/meta-arguments/lifecycle.html
[3]: /docs/providers/aws/r/spot_instance_request.html

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import launch configurations using the `name`. For example:

```terraform
import {
  to = aws_launch_configuration.as_conf
  id = "terraform-lg-123456"
}
```

Using `terraform import`, import launch configurations using the `name`. For example:

```console
% terraform import aws_launch_configuration.as_conf terraform-lg-123456
```
