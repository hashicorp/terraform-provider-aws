---
subcategory: "Network Manager"
layout: "aws"
page_title: "AWS: aws_networkmanager_vpc_attachment"
description: |-
  Terraform resource for managing an AWS NetworkManager VpcAttachment.
---

# Resource: aws_networkmanager_vpc_attachment

Terraform resource for managing an AWS NetworkManager VpcAttachment.

## Example Usage

### Basic Usage

```terraform
resource "aws_networkmanager_vpc_attachment" "example" {
  subnet_arns     = [aws_subnet.example.arn]
  core_network_id = awscc_networkmanager_core_network.example.id
  vpc_arn         = aws_vpc.example.arn
}
```

## Argument Reference

The following arguments are required:

* `core_network_id` - (Required) The ID of a core network for the VPC attachment.
* `subnet_arns` - (Required) The subnet ARN of the VPC attachment.
* `vpc_arn` - (Required) The ARN of the VPC.

The following arguments are optional:

* `options` - (Optional) Options for the VPC attachment.
* `tags` - (Optional) Key-value tags for the attachment. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

### options

* `appliance_mode_support` - (Optional) Indicates whether appliance mode is supported. If enabled, traffic flow between a source and destination use the same Availability Zone for the VPC attachment for the lifetime of that flow.
* `ipv6_support` - (Optional) Indicates whether IPv6 is supported.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - The ARN of the attachment.
* `attachment_policy_rule_number` - The policy rule number associated with the attachment.
* `attachment_type` - The type of attachment.
* `core_network_arn` - The ARN of a core network.
* `edge_location` - The Region where the edge is located.
* `id` - The ID of the attachment.
* `owner_account_id` - The ID of the attachment account owner.
* `resource_arn` - The attachment resource ARN.
* `segment_name` - The name of the segment attachment.
* `state` - The state of the attachment.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import `aws_networkmanager_vpc_attachment` using the attachment ID. For example:

```terraform
import {
  to = aws_networkmanager_vpc_attachment.example
  id = "attachment-0f8fa60d2238d1bd8"
}
```

Using `terraform import`, import `aws_networkmanager_vpc_attachment` using the attachment ID. For example:

```console
% terraform import aws_networkmanager_vpc_attachment.example attachment-0f8fa60d2238d1bd8
```
