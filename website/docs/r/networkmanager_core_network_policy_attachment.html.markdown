---
subcategory: "Network Manager"
layout: "aws"
page_title: "AWS: aws_networkmanager_core_network_policy_attachment"
description: |-
  Provides a Core Network Policy Attachment resource.
---

# Resource: aws_networkmanager_core_network_policy_attachment

Provides a Core Network Policy Attachment resource. This puts a Core Network Policy to an existing Core Network and executes the change set, which deploys changes globally based on the policy submitted (Sets the policy to `LIVE`).

~> **NOTE on Core Networks and Policy Attachments:** For a given policy attachment, this resource is incompatible with using the [`aws_networkmanager_core_network` resource](/docs/providers/aws/r/networkmanager_core_network.html) `policy_document` argument. When using that argument and this resource, both will attempt to manage the core network's policy document and Terraform will show a permanent difference.

~> **NOTE:** Deleting this resource will not delete the current policy defined in this resource. Deleting this resource will also not revert the current `LIVE` policy to the previous version.

## Example Usage

### Basic

```terraform
resource "aws_networkmanager_core_network" "example" {
  global_network_id = aws_networkmanager_global_network.example.id
}

resource "aws_networkmanager_core_network_policy_attachment" "example" {
  core_network_id = aws_networkmanager_core_network.example.id
  policy_document = data.aws_networkmanager_core_network_policy_document.example.json
}
```

### With VPC Attachment

The example below illustrates the scenario where your policy document has static routes pointing to VPC attachments and you want to attach your VPCs to the core network before applying the desired policy document. Set the `create_base_policy` argument of the [`aws_networkmanager_core_network` resource](/docs/providers/aws/r/networkmanager_core_network.html) to `true` if your core network does not currently have any `LIVE` policies (e.g. this is the first `terraform apply` with the core network resource), since a `LIVE` policy is required before VPCs can be attached to the core network. Otherwise, if your core network already has a `LIVE` policy, you may exclude the `create_base_policy` argument.

```terraform
resource "aws_networkmanager_global_network" "example" {}

data "aws_networkmanager_core_network_policy_document" "example" {
  core_network_configuration {
    asn_ranges = ["65022-65534"]

    edge_locations {
      location = "us-west-2"
    }
  }

  segments {
    name = "segment"
  }

  segment_actions {
    action  = "create-route"
    segment = "segment"
    destination_cidr_blocks = [
      "0.0.0.0/0"
    ]
    destinations = [
      aws_networkmanager_vpc_attachment.example.id,
    ]
  }
}

resource "aws_networkmanager_core_network" "example" {
  global_network_id  = aws_networkmanager_global_network.example.id
  create_base_policy = true
}

resource "aws_networkmanager_core_network_policy_attachment" "example" {
  core_network_id = aws_networkmanager_core_network.example.id
  policy_document = data.aws_networkmanager_core_network_policy_document.example.json
}

resource "aws_networkmanager_vpc_attachment" "example" {
  core_network_id = aws_networkmanager_core_network.example.id
  subnet_arns     = aws_subnet.example[*].arn
  vpc_arn         = aws_vpc.example.arn
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
