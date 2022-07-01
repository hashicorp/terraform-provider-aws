---
subcategory: "EC2 (Elastic Compute Cloud)"
layout: "aws"
page_title: "AWS: aws_ec2_host"
description: |-
  Provides an EC2 Host resource. This allows Dedicated Hosts to be allocated, modified, and released.
---

# Resource: aws_ec2_host

Provides an EC2 Host resource. This allows Dedicated Hosts to be allocated, modified, and released.

## Example Usage

```terraform
# Create a new host with instance type of c5.18xlarge with Auto Placement
# and Host Recovery enabled.
resource "aws_ec2_host" "test" {
  instance_type     = "c5.18xlarge"
  availability_zone = "us-west-2a"
  host_recovery     = "on"
  auto_placement    = "on"
}
```

## Argument Reference

The following arguments are supported:

* `auto_placement` - (Optional) Indicates whether the host accepts any untargeted instance launches that match its instance type configuration, or if it only accepts Host tenancy instance launches that specify its unique host ID. Valid values: `on`, `off`. Default: `on`.
* `availability_zone` - (Required) The Availability Zone in which to allocate the Dedicated Host.
* `host_recovery` - (Optional) Indicates whether to enable or disable host recovery for the Dedicated Host. Valid values: `on`, `off`. Default: `off`.
* `instance_family` - (Optional) Specifies the instance family to be supported by the Dedicated Hosts. If you specify an instance family, the Dedicated Hosts support multiple instance types within that instance family. Exactly one of `instance_family` or `instance_type` must be specified.
* `instance_type` - (Optional) Specifies the instance type to be supported by the Dedicated Hosts. If you specify an instance type, the Dedicated Hosts support instances of the specified instance type only. Exactly one of `instance_family` or `instance_type` must be specified.
* `outpost_arn` - (Optional) The Amazon Resource Name (ARN) of the AWS Outpost on which to allocate the Dedicated Host.
* `tags` - (Optional) Map of tags to assign to this resource. If configured with a provider [`default_tags` configuration block](https://www.terraform.io/docs/providers/aws/index.html#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The ID of the allocated Dedicated Host. This is used to launch an instance onto a specific host.
* `arn` - The ARN of the Dedicated Host.
* `owner_id` - The ID of the AWS account that owns the Dedicated Host.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://www.terraform.io/docs/providers/aws/index.html#default_tags-configuration-block).

## Import

Hosts can be imported using the host `id`, e.g.,

```
$ terraform import aws_ec2_host.example h-0385a99d0e4b20cbb
```
