---
subcategory: "Network Manager"
layout: "aws"
page_title: "AWS: aws_networkmanager_transit_gateway_registration"
description: |-
  Registers a transit gateway to a global network.
---

# Resource: aws_networkmanager_transit_gateway_registration

Registers a transit gateway to a global network. The transit gateway can be in any AWS Region,
but it must be owned by the same AWS account that owns the global network.
You cannot register a transit gateway in more than one global network.

## Example Usage

```terraform
resource "aws_networkmanager_global_network" "example" {
  description = "example"
}

resource "aws_ec2_transit_gateway" "example" {}

resource "aws_networkmanager_transit_gateway_registration" "example" {
  global_network_id   = aws_networkmanager_global_network.example.id
  transit_gateway_arn = aws_ec2_transit_gateway.example.arn
}
```

## Argument Reference

This resource supports the following arguments:

* `global_network_id` - (Required) The ID of the Global Network to register to.
* `transit_gateway_arn` - (Required) The ARN of the Transit Gateway to register.

## Attribute Reference

This resource exports no additional attributes.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import `aws_networkmanager_transit_gateway_registration` using the global network ID and transit gateway ARN. For example:

```terraform
import {
  to = aws_networkmanager_transit_gateway_registration.example
  id = "global-network-0d47f6t230mz46dy4,arn:aws:ec2:us-west-2:123456789012:transit-gateway/tgw-123abc05e04123abc"
}
```

Using `terraform import`, import `aws_networkmanager_transit_gateway_registration` using the global network ID and transit gateway ARN. For example:

```console
% terraform import aws_networkmanager_transit_gateway_registration.example global-network-0d47f6t230mz46dy4,arn:aws:ec2:us-west-2:123456789012:transit-gateway/tgw-123abc05e04123abc
```
