---
subcategory: "Direct Connect"
layout: "aws"
page_title: "AWS: aws_dx_hosted_transit_virtual_interface_accepter"
description: |-
  Provides a resource to manage the accepter's side of a Direct Connect hosted transit virtual interface.
---

# Resource: aws_dx_hosted_transit_virtual_interface_accepter

Provides a resource to manage the accepter's side of a Direct Connect hosted transit virtual interface.
This resource accepts ownership of a transit virtual interface created by another AWS account.

-> **NOTE:** AWS allows a Direct Connect hosted transit virtual interface to be deleted from either the allocator's or accepter's side. However, Terraform only allows the Direct Connect hosted transit virtual interface to be deleted from the allocator's side by removing the corresponding `aws_dx_hosted_transit_virtual_interface` resource from your configuration. Removing a `aws_dx_hosted_transit_virtual_interface_accepter` resource from your configuration will remove it from your statefile and management, **but will not delete the Direct Connect virtual interface.**

## Example Usage

```terraform
provider "aws" {
  # Creator's credentials.
}

provider "aws" {
  alias = "accepter"

  # Accepter's credentials.
}

data "aws_caller_identity" "accepter" {
  provider = aws.accepter
}

# Creator's side of the VIF
resource "aws_dx_hosted_transit_virtual_interface" "creator" {
  connection_id    = "dxcon-zzzzzzzz"
  owner_account_id = data.aws_caller_identity.accepter.account_id

  name           = "tf-transit-vif-example"
  vlan           = 4094
  address_family = "ipv4"
  bgp_asn        = 65352

  # The aws_dx_hosted_transit_virtual_interface
  # must be destroyed before the aws_dx_gateway.
  depends_on = [aws_dx_gateway.example]
}

# Accepter's side of the VIF.
resource "aws_dx_gateway" "example" {
  provider = aws.accepter

  name            = "tf-dxg-example"
  amazon_side_asn = 64512
}

resource "aws_dx_hosted_transit_virtual_interface_accepter" "accepter" {
  provider             = aws.accepter
  virtual_interface_id = aws_dx_hosted_transit_virtual_interface.creator.id
  dx_gateway_id        = aws_dx_gateway.example.id

  tags = {
    Side = "Accepter"
  }
}
```

## Argument Reference

This resource supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `dx_gateway_id` - (Required) The ID of the [Direct Connect gateway](dx_gateway.html) to which to connect the virtual interface.
* `virtual_interface_id` - (Required) The ID of the Direct Connect virtual interface to accept.
* `tags` - (Optional) A map of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - The ID of the virtual interface.
* `arn` - The ARN of the virtual interface.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

- `create` - (Default `10m`)
- `delete` - (Default `10m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Direct Connect hosted transit virtual interfaces using the VIF `id`. For example:

```terraform
import {
  to = aws_dx_hosted_transit_virtual_interface_accepter.test
  id = "dxvif-33cc44dd"
}
```

Using `terraform import`, import Direct Connect hosted transit virtual interfaces using the VIF `id`. For example:

```console
% terraform import aws_dx_hosted_transit_virtual_interface_accepter.test dxvif-33cc44dd
```
