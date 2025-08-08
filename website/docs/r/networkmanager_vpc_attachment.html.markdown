---
subcategory: "Network Manager"
layout: "aws"
page_title: "AWS: aws_networkmanager_vpc_attachment"
description: |-
  Manages a Network Manager VPC attachment.
---

# Resource: aws_networkmanager_vpc_attachment

Manages a Network Manager VPC attachment.

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

* `core_network_id` - (Required) ID of a core network for the VPC attachment.
* `subnet_arns` - (Required) Subnet ARNs of the VPC attachment.
* `vpc_arn` - (Required) ARN of the VPC.

The following arguments are optional:

* `options` - (Optional) Options for the VPC attachment. [See below](#options).
* `tags` - (Optional) Key-value tags for the attachment. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

### options

* `appliance_mode_support` - (Optional) Whether to enable appliance mode support. If enabled, traffic flow between a source and destination use the same Availability Zone for the VPC attachment for the lifetime of that flow. If the VPC attachment is pending acceptance, changing this value will recreate the resource.
* `ipv6_support` - (Optional) Whether to enable IPv6 support. If the VPC attachment is pending acceptance, changing this value will recreate the resource.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the attachment.
* `attachment_policy_rule_number` - Policy rule number associated with the attachment.
* `attachment_type` - Type of attachment.
* `core_network_arn` - ARN of a core network.
* `edge_location` - Region where the edge is located.
* `id` - ID of the attachment.
* `owner_account_id` - ID of the attachment account owner.
* `resource_arn` - Attachment resource ARN.
* `segment_name` - Name of the segment attachment.
* `state` - State of the attachment.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `15m`)
* `delete` - (Default `10m`)
* `update` - (Default `10m`)

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
