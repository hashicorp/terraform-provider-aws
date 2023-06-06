---
subcategory: "Device Farm"
layout: "aws"
page_title: "AWS: aws_devicefarm_network_profile"
description: |-
  Provides a Devicefarm network profile
---

# Resource: aws_devicefarm_network_profile

Provides a resource to manage AWS Device Farm Network Profiles.
∂
~> **NOTE:** AWS currently has limited regional support for Device Farm (e.g., `us-west-2`). See [AWS Device Farm endpoints and quotas](https://docs.aws.amazon.com/general/latest/gr/devicefarm.html) for information on supported regions.

## Example Usage

```terraform
resource "aws_devicefarm_project" "example" {
  name = "example"
}

resource "aws_devicefarm_network_profile" "example" {
  name        = "example"
  project_arn = aws_devicefarm_project.example.arn
}
```

## Argument Reference

* `description` - (Optional) The description of the network profile.
* `downlink_bandwidth_bits` - (Optional) The data throughput rate in bits per second, as an integer from `0` to `104857600`. Default value is `104857600`.
* `downlink_delay_ms` - (Optional) Delay time for all packets to destination in milliseconds as an integer from `0` to `2000`.
* `downlink_jitter_ms` - (Optional) Time variation in the delay of received packets in milliseconds as an integer from `0` to `2000`.
* `downlink_loss_percent` - (Optional) Proportion of received packets that fail to arrive from `0` to `100` percent.
* `name` - (Required) The name for the network profile.
* `uplink_bandwidth_bits` - (Optional) The data throughput rate in bits per second, as an integer from `0` to `104857600`. Default value is `104857600`.
* `uplink_delay_ms` - (Optional) Delay time for all packets to destination in milliseconds as an integer from `0` to `2000`.
* `uplink_jitter_ms` - (Optional) Time variation in the delay of received packets in milliseconds as an integer from `0` to `2000`.
* `uplink_loss_percent` - (Optional) Proportion of received packets that fail to arrive from `0` to `100` percent.
* `project_arn` - (Required) The ARN of the project for the network profile.
* `type` - (Optional) The type of network profile to create. Valid values are listed are `PRIVATE` and `CURATED`.
* `tags` - (Optional) A map of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - The Amazon Resource Name of this network profile.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Import

DeviceFarm Network Profiles can be imported by their arn:

```
$ terraform import aws_devicefarm_network_profile.example arn:aws:devicefarm:us-west-2:123456789012:networkprofile:4fa784c7-ccb4-4dbf-ba4f-02198320daa1
```
