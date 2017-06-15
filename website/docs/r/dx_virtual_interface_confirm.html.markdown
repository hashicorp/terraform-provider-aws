---
layout: "aws"
page_title: "AWS: aws_dx_virtual_interfaec_confirm"
sidebar_current: "docs-aws-resource-dx-virtual-interface-confirm"
description: |-
  Providers a Direct Connect Virtual Interface confirmation resource.
---

# aws\_dx\_virtual\_interface\_confirm

Provides a Direct Connect Virtual Interface confirmation resource.

## Example Usage

Due to the interface possibly being created by an account out of your control
it's advisable to specify `prevent_destroy` in a [lifecycle][1] block.

```
resource "aws_dx_virtual_interface_confirm" "vif" {
  virtual_interface_id = "dxvif-abc123"
  interface_type       = "public"

  lifecycle {
    prevent_destroy = true
  }
}
```

## Argument Reference

The following arguments are supported:

* `virtual_interface_id` - (Required) The ID of the virtual interface.
* `interface_type` - (Required) The type of virtual interface, either `public` or `private`.
* `virtual_gateway_id` - (Required only when `interface_type` is "private") The ID of a virtual gateway to attach the interface to.
* `allow_down_state` - (Optional) Whether to allow the virtual interface to be BGP down.

## Attributes Reference

The following attributes are exported:

* `connection_id` - The ID of the connection.
* `asn` - The autonomous system (AS) number for the Border Gateway Protocol (BGP) configuration.
* `virtual_interface_name` - The name of the virtual interface assigned by the customer.
* `vlan` - The VLAN ID.
* `amazon_address` - IP address assigned to the Amazon interface.
* `customer_address` - IP address assigned to the customer interface.
* `owner_account_id` - The AWS account that will own the new virtual interface.
* `auth_key` - The authentication key for BGP configuration.
* `route_filter_prefixes` - A list of routes to be advertised to the AWS network in this region.

[1]: /docs/configuration/resources.html#lifecycle

## Import

Direct Connect Virtual Interfaces can be imported using the `virtual_interface_id`, e.g.

```
$ terraform import aws_dx_virtual_interface_confirm.vif dxvif-abc123
```
