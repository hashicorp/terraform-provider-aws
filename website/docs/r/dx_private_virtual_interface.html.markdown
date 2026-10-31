---
subcategory: "Direct Connect"
layout: "aws"
page_title: "AWS: aws_dx_private_virtual_interface"
description: |-
  Provides a Direct Connect private virtual interface resource.
---

# Resource: aws_dx_private_virtual_interface

Provides a Direct Connect private virtual interface resource.

## Example Usage

```terraform
resource "aws_dx_private_virtual_interface" "foo" {
  connection_id = "dxcon-zzzzzzzz"

  name           = "vif-foo"
  vlan           = 4094
  address_family = "ipv4"
  bgp_asn        = 65352
}
```

## Argument Reference

This resource supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `address_family` - (Required) The address family for the BGP peer. `ipv4 ` or `ipv6`.
* `bgp_asn` - (Optional) BGP autonomous system number as an integer between `1` and `2147483646`. For larger values, use `bgp_asn_long`. Exactly one of `bgp_asn` or `bgp_asn_long` must be specified.
* `bgp_asn_long` - (Optional) BGP autonomous system number as an asplain decimal string between `1` and `4294967294`. This argument also accepts values in the `bgp_asn` range. Exactly one of `bgp_asn` or `bgp_asn_long` must be specified.
* `connection_id` - (Required) The ID of the Direct Connect connection (or LAG) on which to create the virtual interface.
* `name` - (Required) The name for the virtual interface.
* `vlan` - (Required) The VLAN ID.
* `amazon_address` - (Optional) The IPv4 CIDR address to use to send traffic to Amazon. Required for IPv4 BGP peers.
* `bgp_auth_key` - (Optional) Authentication key for BGP configuration. Terraform marks this value as sensitive and redacts it from normal output, but configured or Direct Connect-generated values remain stored in Terraform state. References to this attribute propagate sensitivity, so outputs exposing it must also be marked sensitive. Users must secure and restrict access to Terraform state.
* `customer_address` - (Optional) The IPv4 CIDR destination address to which Amazon should send traffic. Required for IPv4 BGP peers.
* `dx_gateway_id` - (Optional) The ID of the Direct Connect gateway to which to connect the virtual interface.
* `mtu` - (Optional) The maximum transmission unit (MTU) is the size, in bytes, of the largest permissible packet that can be passed over the connection.
The MTU of a virtual private interface can be either `1500` or `9001` (jumbo frames). Default is `1500`.
* `prefix_pool_allocated_count_ipv4` - (Optional) The number of inbound IPv4 route prefixes to allocate to the virtual interface. Valid values are `0` to `1000`. If not specified, AWS applies the default allocation of `100`.
* `prefix_pool_allocated_count_ipv6` - (Optional) The number of inbound IPv6 route prefixes to allocate to the virtual interface. Valid values are `0` to `1000`. If not specified, AWS applies the default allocation of `100`.
* `rate_limit` - (Optional) Maximum bandwidth allocation for the virtual interface, restricting the bandwidth it can use on the parent connection. Specify a supported bandwidth value without a space (for example, `50Mbps`, `1Gbps`, or `10Gbps`); the value cannot exceed the bandwidth of the parent connection or link aggregation group (LAG), and supported values range up to `1.6Tbps`. See the [VIF Rate Limiters documentation](https://docs.aws.amazon.com/directconnect/latest/UserGuide/vif-rate-limiters.html) for the full list of supported values. Rate Limiters are supported only on Direct Connect dedicated connections (including LAGs); they are not supported on hosted connections.
* `sitelink_enabled` - (Optional) Indicates whether to enable or disable SiteLink.
* `tags` - (Optional) A map of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.
* `vpn_gateway_id` - (Optional) The ID of the [virtual private gateway](vpn_gateway.html) to which to connect the virtual interface.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - The ID of the virtual interface.
* `arn` - The ARN of the virtual interface.
* `aws_device` - The Direct Connect endpoint on which the virtual interface terminates.
* `jumbo_frame_capable` - Indicates whether jumbo frames (9001 MTU) are supported.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

- `create` - (Default `10m`)
- `update` - (Default `10m`)
- `delete` - (Default `10m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Direct Connect private virtual interfaces using the VIF `id`. For example:

```terraform
import {
  to = aws_dx_private_virtual_interface.test
  id = "dxvif-33cc44dd"
}
```

Using `terraform import`, import Direct Connect private virtual interfaces using the VIF `id`. For example:

```console
% terraform import aws_dx_private_virtual_interface.test dxvif-33cc44dd
```

~> **Note:** When a virtual interface uses an ASN in the `bgp_asn` range (`1` to `2147483646`), AWS returns the value in both the `asn` and `asnLong` API fields, so import always populates `bgp_asn` rather than `bgp_asn_long`. If the virtual interface was originally created with `bgp_asn_long` set to a value in that range, update your configuration to use `bgp_asn` after import to avoid a difference. Virtual interfaces using a 4-byte ASN (greater than `2147483646`) import into `bgp_asn_long` as expected.
