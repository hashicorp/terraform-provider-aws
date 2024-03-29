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

This resource supports the following arguments:

* `asset_id` - (Optional) The ID of the Outpost hardware asset on which to allocate the Dedicated Hosts. This parameter is supported only if you specify OutpostArn. If you are allocating the Dedicated Hosts in a Region, omit this parameter.
* `auto_placement` - (Optional) Indicates whether the host accepts any untargeted instance launches that match its instance type configuration, or if it only accepts Host tenancy instance launches that specify its unique host ID. Valid values: `on`, `off`. Default: `on`.
* `availability_zone` - (Required) The Availability Zone in which to allocate the Dedicated Host.
* `host_recovery` - (Optional) Indicates whether to enable or disable host recovery for the Dedicated Host. Valid values: `on`, `off`. Default: `off`.
* `instance_family` - (Optional) Specifies the instance family to be supported by the Dedicated Hosts. If you specify an instance family, the Dedicated Hosts support multiple instance types within that instance family. Exactly one of `instance_family` or `instance_type` must be specified.
* `instance_type` - (Optional) Specifies the instance type to be supported by the Dedicated Hosts. If you specify an instance type, the Dedicated Hosts support instances of the specified instance type only. Exactly one of `instance_family` or `instance_type` must be specified.
* `outpost_arn` - (Optional) The Amazon Resource Name (ARN) of the AWS Outpost on which to allocate the Dedicated Host.
* `tags` - (Optional) Map of tags to assign to this resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - The ID of the allocated Dedicated Host. This is used to launch an instance onto a specific host.
* `arn` - The ARN of the Dedicated Host.
* `owner_id` - The ID of the AWS account that owns the Dedicated Host.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `10m`)
* `update` - (Default `10m`)
* `delete` - (Default `20m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import hosts using the host `id`. For example:

```terraform
import {
  to = aws_ec2_host.example
  id = "h-0385a99d0e4b20cbb"
}
```

Using `terraform import`, import hosts using the host `id`. For example:

```console
% terraform import aws_ec2_host.example h-0385a99d0e4b20cbb
```
