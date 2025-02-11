---
subcategory: "Elemental MediaPackage Version 2"
layout: "aws"
page_title: "AWS: aws_media_packagev2_channel_group"
description: |-
  Creates an AWS Elemental MediaPackage Version 2 Channel Group.
---

# Resource: aws_media_packagev2_channel_group

Creates an AWS Elemental MediaPackage Version 2 Channel Group.

## Example Usage

```terraform
resource "aws_media_packagev2_channel_group" "example" {
  name        = "example"
  description = "channel group for example channels"
}
```

## Argument Reference

This resource supports the following arguments:

* `name` - (Required) A unique identifier naming the channel group
* `description` - (Optional) A description of the channel group
* `tags` - (Optional) A map of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - The ARN of the channel
* `description` - The same as `description`
* `egress_domain` - The egress domain of the channel group
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import an Elemental MediaPackage Version 2 Channel Group using the channel group's name. For example:

```terraform
import {
  to = aws_media_packagev2_channel_group.example
  id = "example"
}
```

Using `terraform import`, import Elemental MediaPackage Version 2 Channel Group using the channel group's `name`. For example:

```console
% terraform import aws_media_packagev2_channel_group.example example
```
