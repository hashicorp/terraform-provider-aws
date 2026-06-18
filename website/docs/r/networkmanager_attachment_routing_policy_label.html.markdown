---
subcategory: "Network Manager"
layout: "aws"
page_title: "AWS: aws_networkmanager_attachment_routing_policy_label"
description: |-
  Associates a routing policy label to a Network Manager Cloud WAN's attachment.
---

# Resource: aws_networkmanager_attachment_routing_policy_label

Associates a routing policy label to a Network Manager Cloud WAN's attachment outside of the attachment creation. This is useful in multi-account environments where only the Cloud WAN core network owner account can apply a routing policy label.

## Example Usage

```terraform
resource "aws_networkmanager_attachment_routing_policy_label" "example" {
  core_network_id      = aws_networkmanager_core_network.example.id
  attachment_id        = aws_networkmanager_vpc_attachment.example.id
  routing_policy_label = "attachmentPolicyLabel"
}
```

## Argument Reference

The following arguments are required:

* `attachment_id` - (Required, Forces new resource) The ID of the attachment to apply the routing policy label to.
* `core_network_id` - (Required, Forces new resource) The ID of the core network containing the attachment.
* `routing_policy_label` - (Required, Forces new resource) The routing policy label to apply to the attachment.

## Attribute Reference

This resource exports no additional attributes.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import `aws_networkmanager_attachment_routing_policy_label` using the core network ID and attachment ID separated by a comma (`,`). For example:

```terraform
import {
  to = aws_networkmanager_attachment_routing_policy_label.example
  id = "core-network-0fab1c1e1e1e1e1e1,attachment-0fab2c2e2e2e2e2e2"
}
```

Using `terraform import`, import `aws_networkmanager_attachment_routing_policy_label` using the core network ID and attachment ID separated by a comma (`,`). For example:

```console
% terraform import aws_networkmanager_attachment_routing_policy_label.example core-network-0fab1c1e1e1e1e1e1,attachment-0fab2c2e2e2e2e2e2
```
