---
subcategory: "Direct Connect"
layout: "aws"
page_title: "AWS: aws_dx_connection"
description: |-
  Provides a Connection of Direct Connect.
---

# Resource: aws_dx_connection

Provides a Connection of Direct Connect.

## Example Usage

### Create a connection

```terraform
resource "aws_dx_connection" "hoge" {
  name      = "tf-dx-connection"
  bandwidth = "1Gbps"
  location  = "EqDC2"
}
```

### Request a MACsec-capable connection

```terraform
resource "aws_dx_connection" "example" {
  name           = "tf-dx-connection"
  bandwidth      = "10Gbps"
  location       = "EqDA2"
  request_macsec = true
}
```

### Configure encryption mode for MACsec-capable connections

-> **NOTE:** You can only specify the `encryption_mode` argument once the connection is in an `Available` state.

```terraform
resource "aws_dx_connection" "example" {
  name            = "tf-dx-connection"
  bandwidth       = "10Gbps"
  location        = "EqDC2"
  request_macsec  = true
  encryption_mode = "must_encrypt"
}
```

## Argument Reference

This resource supports the following arguments:

* `bandwidth` - (Required) The bandwidth of the connection. Valid values for dedicated connections: 1Gbps, 10Gbps, 100Gbps, and 400Gbps. Valid values for hosted connections: 50Mbps, 100Mbps, 200Mbps, 300Mbps, 400Mbps, 500Mbps, 1Gbps, 2Gbps, 5Gbps, 10Gbps, and 25Gbps. Case sensitive. Refer to the AWS Direct Connection supported bandwidths for [Dedicated Connections](https://docs.aws.amazon.com/directconnect/latest/UserGuide/dedicated_connection.html) and [Hosted Connections](https://docs.aws.amazon.com/directconnect/latest/UserGuide/hosted_connection.html).
* `encryption_mode` - (Optional) The connection MAC Security (MACsec) encryption mode. MAC Security (MACsec) is only available on dedicated connections. Valid values are `no_encrypt`, `should_encrypt`, and `must_encrypt`.
* `location` - (Required) The AWS Direct Connect location where the connection is located. See [DescribeLocations](https://docs.aws.amazon.com/directconnect/latest/APIReference/API_DescribeLocations.html) for the list of AWS Direct Connect locations. Use `locationCode`.
* `name` - (Required) The name of the connection.
* `provider_name` - (Optional) The name of the service provider associated with the connection.
* `request_macsec` - (Optional) Boolean value indicating whether you want the connection to support MAC Security (MACsec). MAC Security (MACsec) is only available on dedicated connections. See [MACsec prerequisites](https://docs.aws.amazon.com/directconnect/latest/UserGuide/direct-connect-mac-sec-getting-started.html#mac-sec-prerequisites) for more information about MAC Security (MACsec) prerequisites. Default value: `false`.

~> **NOTE:** Changing the value of `request_macsec` will cause the resource to be destroyed and re-created.

* `skip_destroy` - (Optional) Set to true if you do not wish the connection to be deleted at destroy time, and instead just removed from the Terraform state.
* `tags` - (Optional) A map of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - The ARN of the connection.
* `aws_device` - The Direct Connect endpoint on which the physical connection terminates.
* `has_logical_redundancy` - Indicates whether the connection supports a secondary BGP peer in the same address family (IPv4/IPv6).
* `id` - The ID of the connection.
* `jumbo_frame_capable` - Boolean value representing if jumbo frames have been enabled for this connection.
* `macsec_capable` - Boolean value indicating whether the connection supports MAC Security (MACsec).
* `owner_account_id` - The ID of the AWS account that owns the connection.
* `partner_name` - The name of the AWS Direct Connect service provider associated with the connection.
* `port_encryption_status` - The MAC Security (MACsec) port link status of the connection.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).
* `vlan_id` - The VLAN ID.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Direct Connect connections using the connection `id`. For example:

```terraform
import {
  to = aws_dx_connection.test_connection
  id = "dxcon-ffre0ec3"
}
```

Using `terraform import`, import Direct Connect connections using the connection `id`. For example:

```console
% terraform import aws_dx_connection.test_connection dxcon-ffre0ec3
```
