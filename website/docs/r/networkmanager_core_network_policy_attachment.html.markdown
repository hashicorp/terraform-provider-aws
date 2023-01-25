---
subcategory: "Network Manager"
layout: "aws"
page_title: "AWS: aws_networkmanager_core_network_policy_attachment"
description: |-
  Provides a Core Network Policy Attachment resource.
---

# Resource: aws_networkmanager_core_network_policy_attachment

Provides a Core Network Policy Attachment resource.

~> **NOTE on Core Networks and Policy Attachments:** Terraform currently provides both a standalone [`aws_networkmanager_core_network_policy_attachment`](networkmanager_core_network_policy_attachment.html) resource, and an [`aws_networkmanager_core_network`](networkmanager_core_network.html) resource with `policy_document` defined in-line. These two methods are not mutually-exclusive. If `aws_networkmanager_core_network_policy_attachment` resources are used with inline `policy_document`, the `aws_networkmanager_core_network` resource must be configured to ignore changes to the `policy_document` argument within a [`lifecycle` configuration block](https://www.terraform.io/docs/configuration/meta-arguments/lifecycle.html).


## Example Usage

### Basic

```terraform
resource "aws_networkmanager_core_network" "example" {
  global_network_id = aws_networkmanager_global_network.example.id

  lifecycle {
    ignore_changes = [policy_document]
  }
}

resource "aws_networkmanager_core_network_policy_attachment" "example" {
  core_network_id = aws_networkmanager_core_network.example.id
  policy_document = data.aws_networkmanager_core_network_policy_document.example.json
}
```

## Argument Reference

The following arguments are supported:

* `core_network_id` - (Required) The ID of the core network that a policy will be attached to and made `LIVE`.
* `policy_document` - (Required) Policy document for creating a core network. Note that updating this argument will result in the new policy document version being set as the `LATEST` and `LIVE` policy document. Refer to the [Core network policies documentation](https://docs.aws.amazon.com/network-manager/latest/cloudwan/cloudwan-policy-change-sets.html) for more information.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `update` - (Default `30m`). If this is the first time attaching a policy to a core network then this timeout value is also used as the `create` timeout value.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `state` - Current state of a core network.

## Import

`aws_networkmanager_core_network_policy_attachment` can be imported using the core network ID, e.g.

```
$ terraform import aws_networkmanager_core_network_policy_attachment.example core-network-0d47f6t230mz46dy4
```
