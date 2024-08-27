---
subcategory: "Network Manager"
layout: "aws"
page_title: "AWS: aws_networkmanager_connect_attachment"
description: |-
  Terraform resource for managing an AWS Network Manager ConnectAttachment.
---

# Resource: aws_networkmanager_connect_attachment

Terraform resource for managing an AWS Network Manager ConnectAttachment.

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
    "aws_networkmanager_attachment_accepter.test"
  ]
}

resource "aws_networkmanager_attachment_accepter" "example2" {
  attachment_id   = aws_networkmanager_connect_attachment.example.id
  attachment_type = aws_networkmanager_connect_attachment.example.attachment_type
}
```

## Argument Reference

The following arguments are required:

- `core_network_id` - (Required) The ID of a core network where you want to create the attachment.
- `transport_attachment_id` - (Required) The ID of the attachment between the two connections.
- `edge_location` - (Required) The Region where the edge is located.
- `options` - (Required) Options block. See [options](#options) for more information.

The following arguments are optional:

- `tags` - (Optional) Key-value tags for the attachment. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

### options

* `protocol` - (Required) The protocol used for the attachment connection. Possible values are `GRE` and `NO_ENCAP`.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

- `arn` - The ARN of the attachment.
- `attachment_policy_rule_number` - The policy rule number associated with the attachment.
- `attachment_type` - The type of attachment.
- `core_network_arn` - The ARN of a core network.
- `core_network_id` - The ID of a core network
- `edge_location` - The Region where the edge is located.
- `id` - The ID of the attachment.
- `owner_account_id` - The ID of the attachment account owner.
- `resource_arn` - The attachment resource ARN.
- `segment_name` - The name of the segment attachment.
- `state` - The state of the attachment.
- `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

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
