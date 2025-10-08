---
subcategory: "Network Manager"
layout: "aws"
page_title: "AWS: aws_networkmanager_connect_attachment"
description: |-
  Manages an AWS Network Manager Connect Attachment.
---

# Resource: aws_networkmanager_connect_attachment

Manages an AWS Network Manager Connect Attachment.

Use this resource to create a Connect attachment in AWS Network Manager. Connect attachments enable you to connect your on-premises networks to your core network through a VPC or Transit Gateway attachment.

## Example Usage

### Basic Usage

```terraform
resource "aws_networkmanager_vpc_attachment" "example" {
  subnet_arns     = aws_subnet.example[*].arn
  core_network_id = awscc_networkmanager_core_network.example.id
  vpc_arn         = aws_vpc.example.arn
}

resource "aws_networkmanager_connect_attachment" "example" {
  core_network_id         = awscc_networkmanager_core_network.example.id
  transport_attachment_id = aws_networkmanager_vpc_attachment.example.id
  edge_location           = aws_networkmanager_vpc_attachment.example.edge_location
  options {
    protocol = "GRE"
  }
}
```

### Usage with attachment accepter

```terraform
resource "aws_networkmanager_vpc_attachment" "example" {
  subnet_arns     = aws_subnet.example[*].arn
  core_network_id = awscc_networkmanager_core_network.example.id
  vpc_arn         = aws_vpc.example.arn
}

resource "aws_networkmanager_attachment_accepter" "example" {
  attachment_id   = aws_networkmanager_vpc_attachment.example.id
  attachment_type = aws_networkmanager_vpc_attachment.example.attachment_type
}

resource "aws_networkmanager_connect_attachment" "example" {
  core_network_id         = awscc_networkmanager_core_network.example.id
  transport_attachment_id = aws_networkmanager_vpc_attachment.example.id
  edge_location           = aws_networkmanager_vpc_attachment.example.edge_location
  options {
    protocol = "GRE"
  }
  depends_on = [
    aws_networkmanager_attachment_accepter.example
  ]
}

resource "aws_networkmanager_attachment_accepter" "example2" {
  attachment_id   = aws_networkmanager_connect_attachment.example.id
  attachment_type = aws_networkmanager_connect_attachment.example.attachment_type
}
```

## Argument Reference

The following arguments are required:

* `core_network_id` - (Required) ID of a core network where you want to create the attachment.
* `edge_location` - (Required) Region where the edge is located.
* `options` - (Required) Options block. See [options](#options) for more information.
* `transport_attachment_id` - (Required) ID of the attachment between the two connections.

The following arguments are optional:

* `tags` - (Optional) Key-value tags for the attachment. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

### options

* `protocol` - (Optional) Protocol used for the attachment connection. Valid values: `GRE`, `NO_ENCAP`.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the attachment.
* `attachment_id` - ID of the attachment.
* `attachment_policy_rule_number` - Policy rule number associated with the attachment.
* `attachment_type` - Type of attachment.
* `core_network_arn` - ARN of a core network.
* `owner_account_id` - ID of the attachment account owner.
* `resource_arn` - Attachment resource ARN.
* `segment_name` - Name of the segment attachment.
* `state` - State of the attachment.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `10m`)
* `delete` - (Default `10m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import `aws_networkmanager_connect_attachment` using the attachment ID. For example:

```terraform
import {
  to = aws_networkmanager_connect_attachment.example
  id = "attachment-0f8fa60d2238d1bd8"
}
```

Using `terraform import`, import `aws_networkmanager_connect_attachment` using the attachment ID. For example:

```console
% terraform import aws_networkmanager_connect_attachment.example attachment-0f8fa60d2238d1bd8
```
