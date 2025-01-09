---
subcategory: "Network Manager"
layout: "aws"
page_title: "AWS: aws_networkmanager_dx_gateway_attachment"
description: |-
  Terraform resource for managing an AWS Network Manager Direct Connect Gateway Attachment.
---
# Resource: aws_networkmanager_dx_gateway_attachment

Terraform resource for managing an AWS Network Manager Direct Connect (DX) Gateway Attachment.

## Example Usage

### Basic Usage

```terraform
resource "aws_networkmanager_dx_gateway_attachment" "test" {
  core_network_id            = aws_networkmanager_core_network_policy_attachment.test.core_network_id
  direct_connect_gateway_arn = "arn:aws:directconnect::${data.aws_caller_identity.current.account_id}:dx-gateway/${aws_dx_gateway.test.id}"
  edge_locations             = [data.aws_region.current.name]
}
```

## Argument Reference

The following arguments are required:

* `core_network_id` - (Required) ID of the Cloud WAN core network to which the Direct Connect gateway attachment should be attached.
* `direct_connect_gateway_arn` - (Required) ARN of the Direct Connect gateway attachment.
* `edge_locations` - (Required) One or more core network edge locations to associate with the Direct Connect gateway attachment.

The following arguments are optional:

* `tags` - (Optional) Key-value tags for the attachment. If configured with a provider `default_tags` configuration block present, tags with matching keys will overwrite those defined at the provider-level.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `attachment_policy_rule_number` - Policy rule number associated with the attachment.
* `attachment_type` - Type of attachment.
* `core_network_arn` - ARN of the core network for the attachment.
* `id` - The ID of the attachment.
* `owner_account_id` - ID of the attachment account owner.
* `segment_name` - Name of the segment attachment.
* `state` - State of the attachment.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `30m`)
* `update` - (Default `30m`)
* `delete` - (Default `30m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Network Manager DX Gateway Attachment using the `id`. For example:

```terraform
import {
  to = aws_networkmanager_dx_gateway_attachment.example
  id = "attachment-1a2b3c4d5e6f7g"
}
```

Using `terraform import`, import Network Manager DX Gateway Attachment using the `id`. For example:

```console
% terraform import aws_networkmanager_dx_gateway_attachment.example attachment-1a2b3c4d5e6f7g
```
