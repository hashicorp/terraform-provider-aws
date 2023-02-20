---
subcategory: "FMS (Firewall Manager)"
layout: "aws"
page_title: "AWS: aws_fms_protocol"
description: |-
  Provides a resource to create an AWS Firewall Manager protocol list
---

# Resource: aws_fms_protocol

Provides a resource to create an AWS Firewall Manager protocol list. You need to be using AWS organizations and have enabled the Firewall Manager administrator account.

## Example Usage

```terraform

resource "aws_fms_protocol" "example" {
  name      = "FMS-Protocol-Example"
  protocols = ["IPv4", "IPv6", "ICMP"]
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The friendly name of the AWS Firewall Manager Protocol list.
* `protocols` - (Required) A list of protocols in the AWS Firewall Manager protocol list.
* `tags` - (Optional) Key-value mapping of resource tags. If configured with a provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The ID of the AWS Firewall Manager protocol list.
* `arn` - The ARN of the AWS Firewall Manager protocol list.
* `protocol_update_token` - A unique identifier for each update to the protocol list.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block).

## Import

Firewall Manager protocol list can be imported using the protocol list ID, e.g.,

```
$ terraform import aws_fms_protocol.example 5be49585-a7e3-4c49-dde1-a179fe4a619a
```
