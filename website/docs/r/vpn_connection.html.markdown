---
subcategory: "VPN (Site-to-Site)"
layout: "aws"
page_title: "AWS: aws_vpn_connection"
description: |-
  Manages a Site-to-Site VPN connection. A Site-to-Site VPN connection is an Internet Protocol security (IPsec) VPN connection between a VPC and an on-premises network.
---

# Resource: aws_vpn_connection

Manages a Site-to-Site VPN connection. A Site-to-Site VPN connection is an Internet Protocol security (IPsec) VPN connection between a VPC and an on-premises network.
Any new Site-to-Site VPN connection that you create is an [AWS VPN connection](https://docs.aws.amazon.com/vpn/latest/s2svpn/vpn-categories.html).

~> **Note:** All arguments including `tunnel1_preshared_key` and `tunnel2_preshared_key` will be stored in the raw state as plain-text.
[Read more about sensitive data in state](https://www.terraform.io/docs/state/sensitive-data.html).

~> **Note:** The CIDR blocks in the arguments `tunnel1_inside_cidr` and `tunnel2_inside_cidr` must have a prefix of /30 and be a part of a specific range.
[Read more about this in the AWS documentation](https://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_VpnTunnelOptionsSpecification.html).

## Example Usage

### EC2 Transit Gateway

```terraform
resource "aws_ec2_transit_gateway" "example" {}

resource "aws_customer_gateway" "example" {
  bgp_asn    = 65000
  ip_address = "172.0.0.1"
  type       = "ipsec.1"
}

resource "aws_vpn_connection" "example" {
  customer_gateway_id = aws_customer_gateway.example.id
  transit_gateway_id  = aws_ec2_transit_gateway.example.id
  type                = aws_customer_gateway.example.type
}
```

### Virtual Private Gateway

```terraform
resource "aws_vpc" "vpc" {
  cidr_block = "10.0.0.0/16"
}

resource "aws_vpn_gateway" "vpn_gateway" {
  vpc_id = aws_vpc.vpc.id
}

resource "aws_customer_gateway" "customer_gateway" {
  bgp_asn    = 65000
  ip_address = "172.0.0.1"
  type       = "ipsec.1"
}

resource "aws_vpn_connection" "main" {
  vpn_gateway_id      = aws_vpn_gateway.vpn_gateway.id
  customer_gateway_id = aws_customer_gateway.customer_gateway.id
  type                = "ipsec.1"
  static_routes_only  = true
}
```

### AWS Site to Site Private VPN

```terraform
resource "aws_dx_gateway" "example" {
  name            = "terraform_ipsec_vpn_example"
  amazon_side_asn = "64512"
}

resource "aws_ec2_transit_gateway" "example" {
  amazon_side_asn = "64513"
  description     = "terraform_ipsec_vpn_example"
  transit_gateway_cidr_blocks = [
    "10.0.0.0/24",
  ]
}

resource "aws_customer_gateway" "example" {
  bgp_asn    = 64514
  ip_address = "10.0.0.1"
  type       = "ipsec.1"

  tags = {
    Name = "terraform_ipsec_vpn_example"
  }
}

resource "aws_dx_gateway_association" "example" {
  dx_gateway_id         = aws_dx_gateway.example.id
  associated_gateway_id = aws_ec2_transit_gateway.example.id

  allowed_prefixes = [
    "10.0.0.0/8",
  ]
}

data "aws_ec2_transit_gateway_dx_gateway_attachment" "example" {
  transit_gateway_id = aws_ec2_transit_gateway.example.id
  dx_gateway_id      = aws_dx_gateway.example.id

  depends_on = [
    aws_dx_gateway_association.example
  ]
}

resource "aws_vpn_connection" "example" {
  customer_gateway_id                     = aws_customer_gateway.example.id
  outside_ip_address_type                 = "PrivateIpv4"
  transit_gateway_id                      = aws_ec2_transit_gateway.example.id
  transport_transit_gateway_attachment_id = data.aws_ec2_transit_gateway_dx_gateway_attachment.example.id
  type                                    = "ipsec.1"

  tags = {
    Name = "terraform_ipsec_vpn_example"
  }
}
```

## Argument Reference

This resource supports the following arguments:

* `customer_gateway_id` - (Required) The ID of the customer gateway.
* `type` - (Required) The type of VPN connection. The only type AWS supports at this time is "ipsec.1".
* `transit_gateway_id` - (Optional) The ID of the EC2 Transit Gateway.
* `vpn_gateway_id` - (Optional) The ID of the Virtual Private Gateway.
* `static_routes_only` - (Optional, Default `false`) Whether the VPN connection uses static routes exclusively. Static routes must be used for devices that don't support BGP.
* `enable_acceleration` - (Optional, Default `false`) Indicate whether to enable acceleration for the VPN connection. Supports only EC2 Transit Gateway.
* `tags` - (Optional) Tags to apply to the connection. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.
* `local_ipv4_network_cidr` - (Optional, Default `0.0.0.0/0`) The IPv4 CIDR on the customer gateway (on-premises) side of the VPN connection.
* `local_ipv6_network_cidr` - (Optional, Default `::/0`) The IPv6 CIDR on the customer gateway (on-premises) side of the VPN connection.
* `outside_ip_address_type` - (Optional, Default `PublicIpv4`) Indicates if a Public S2S VPN or Private S2S VPN over AWS Direct Connect. Valid values are `PublicIpv4 | PrivateIpv4`
* `remote_ipv4_network_cidr` - (Optional, Default `0.0.0.0/0`) The IPv4 CIDR on the AWS side of the VPN connection.
* `remote_ipv6_network_cidr` - (Optional, Default `::/0`) The IPv6 CIDR on the AWS side of the VPN connection.
* `transport_transit_gateway_attachment_id` - (Required when outside_ip_address_type is set to `PrivateIpv4`). The attachment ID of the Transit Gateway attachment to Direct Connect Gateway. The ID is obtained through a data source only.
* `tunnel_inside_ip_version` - (Optional, Default `ipv4`) Indicate whether the VPN tunnels process IPv4 or IPv6 traffic. Valid values are `ipv4 | ipv6`. `ipv6` Supports only EC2 Transit Gateway.
* `tunnel1_inside_cidr` - (Optional) The CIDR block of the inside IP addresses for the first VPN tunnel. Valid value is a size /30 CIDR block from the 169.254.0.0/16 range.
* `tunnel2_inside_cidr` - (Optional) The CIDR block of the inside IP addresses for the second VPN tunnel. Valid value is a size /30 CIDR block from the 169.254.0.0/16 range.
* `tunnel1_inside_ipv6_cidr` - (Optional) The range of inside IPv6 addresses for the first VPN tunnel. Supports only EC2 Transit Gateway. Valid value is a size /126 CIDR block from the local fd00::/8 range.
* `tunnel2_inside_ipv6_cidr` - (Optional) The range of inside IPv6 addresses for the second VPN tunnel. Supports only EC2 Transit Gateway. Valid value is a size /126 CIDR block from the local fd00::/8 range.
* `tunnel1_preshared_key` - (Optional) The preshared key of the first VPN tunnel. The preshared key must be between 8 and 64 characters in length and cannot start with zero(0). Allowed characters are alphanumeric characters, periods(.) and underscores(_).
* `tunnel2_preshared_key` - (Optional) The preshared key of the second VPN tunnel. The preshared key must be between 8 and 64 characters in length and cannot start with zero(0). Allowed characters are alphanumeric characters, periods(.) and underscores(_).
* `tunnel1_dpd_timeout_action` - (Optional, Default `clear`) The action to take after DPD timeout occurs for the first VPN tunnel. Specify restart to restart the IKE initiation. Specify clear to end the IKE session. Valid values are `clear | none | restart`.
* `tunnel2_dpd_timeout_action` - (Optional, Default `clear`) The action to take after DPD timeout occurs for the second VPN tunnel. Specify restart to restart the IKE initiation. Specify clear to end the IKE session. Valid values are `clear | none | restart`.
* `tunnel1_dpd_timeout_seconds` - (Optional, Default `30`) The number of seconds after which a DPD timeout occurs for the first VPN tunnel. Valid value is equal or higher than `30`.
* `tunnel2_dpd_timeout_seconds` - (Optional, Default `30`) The number of seconds after which a DPD timeout occurs for the second VPN tunnel. Valid value is equal or higher than `30`.
* `tunnel1_enable_tunnel_lifecycle_control` - (Optional, Default `false`) Turn on or off tunnel endpoint lifecycle control feature for the first VPN tunnel. Valid values are `true | false`.
* `tunnel2_enable_tunnel_lifecycle_control` - (Optional, Default `false`) Turn on or off tunnel endpoint lifecycle control feature for the second VPN tunnel. Valid values are `true | false`.
* `tunnel1_ike_versions` - (Optional) The IKE versions that are permitted for the first VPN tunnel. Valid values are `ikev1 | ikev2`.
* `tunnel2_ike_versions` - (Optional) The IKE versions that are permitted for the second VPN tunnel. Valid values are `ikev1 | ikev2`.
* `tunnel1_log_options` - (Optional) Options for logging VPN tunnel activity. See [Log Options](#log-options) below for more details.
* `tunnel2_log_options` - (Optional) Options for logging VPN tunnel activity. See [Log Options](#log-options) below for more details.
* `tunnel1_phase1_dh_group_numbers` - (Optional) List of one or more Diffie-Hellman group numbers that are permitted for the first VPN tunnel for phase 1 IKE negotiations. Valid values are ` 2 | 14 | 15 | 16 | 17 | 18 | 19 | 20 | 21 | 22 | 23 | 24`.
* `tunnel2_phase1_dh_group_numbers` - (Optional) List of one or more Diffie-Hellman group numbers that are permitted for the second VPN tunnel for phase 1 IKE negotiations. Valid values are ` 2 | 14 | 15 | 16 | 17 | 18 | 19 | 20 | 21 | 22 | 23 | 24`.
* `tunnel1_phase1_encryption_algorithms` - (Optional) List of one or more encryption algorithms that are permitted for the first VPN tunnel for phase 1 IKE negotiations. Valid values are `AES128 | AES256 | AES128-GCM-16 | AES256-GCM-16`.
* `tunnel2_phase1_encryption_algorithms` - (Optional) List of one or more encryption algorithms that are permitted for the second VPN tunnel for phase 1 IKE negotiations. Valid values are `AES128 | AES256 | AES128-GCM-16 | AES256-GCM-16`.
* `tunnel1_phase1_integrity_algorithms` - (Optional) One or more integrity algorithms that are permitted for the first VPN tunnel for phase 1 IKE negotiations. Valid values are `SHA1 | SHA2-256 | SHA2-384 | SHA2-512`.
* `tunnel2_phase1_integrity_algorithms` - (Optional) One or more integrity algorithms that are permitted for the second VPN tunnel for phase 1 IKE negotiations. Valid values are `SHA1 | SHA2-256 | SHA2-384 | SHA2-512`.
* `tunnel1_phase1_lifetime_seconds` - (Optional, Default `28800`) The lifetime for phase 1 of the IKE negotiation for the first VPN tunnel, in seconds. Valid value is between `900` and `28800`.
* `tunnel2_phase1_lifetime_seconds` - (Optional, Default `28800`) The lifetime for phase 1 of the IKE negotiation for the second VPN tunnel, in seconds. Valid value is between `900` and `28800`.
* `tunnel1_phase2_dh_group_numbers` - (Optional) List of one or more Diffie-Hellman group numbers that are permitted for the first VPN tunnel for phase 2 IKE negotiations. Valid values are `2 | 5 | 14 | 15 | 16 | 17 | 18 | 19 | 20 | 21 | 22 | 23 | 24`.
* `tunnel2_phase2_dh_group_numbers` - (Optional) List of one or more Diffie-Hellman group numbers that are permitted for the second VPN tunnel for phase 2 IKE negotiations. Valid values are `2 | 5 | 14 | 15 | 16 | 17 | 18 | 19 | 20 | 21 | 22 | 23 | 24`.
* `tunnel1_phase2_encryption_algorithms` - (Optional) List of one or more encryption algorithms that are permitted for the first VPN tunnel for phase 2 IKE negotiations. Valid values are `AES128 | AES256 | AES128-GCM-16 | AES256-GCM-16`.
* `tunnel2_phase2_encryption_algorithms` - (Optional) List of one or more encryption algorithms that are permitted for the second VPN tunnel for phase 2 IKE negotiations. Valid values are `AES128 | AES256 | AES128-GCM-16 | AES256-GCM-16`.
* `tunnel1_phase2_integrity_algorithms` - (Optional) List of one or more integrity algorithms that are permitted for the first VPN tunnel for phase 2 IKE negotiations. Valid values are `SHA1 | SHA2-256 | SHA2-384 | SHA2-512`.
* `tunnel2_phase2_integrity_algorithms` - (Optional) List of one or more integrity algorithms that are permitted for the second VPN tunnel for phase 2 IKE negotiations. Valid values are `SHA1 | SHA2-256 | SHA2-384 | SHA2-512`.
* `tunnel1_phase2_lifetime_seconds` - (Optional, Default `3600`) The lifetime for phase 2 of the IKE negotiation for the first VPN tunnel, in seconds. Valid value is between `900` and `3600`.
* `tunnel2_phase2_lifetime_seconds` - (Optional, Default `3600`) The lifetime for phase 2 of the IKE negotiation for the second VPN tunnel, in seconds. Valid value is between `900` and `3600`.
* `tunnel1_rekey_fuzz_percentage` - (Optional, Default `100`) The percentage of the rekey window for the first VPN tunnel (determined by `tunnel1_rekey_margin_time_seconds`) during which the rekey time is randomly selected. Valid value is between `0` and `100`.
* `tunnel2_rekey_fuzz_percentage` - (Optional, Default `100`) The percentage of the rekey window for the second VPN tunnel (determined by `tunnel2_rekey_margin_time_seconds`) during which the rekey time is randomly selected. Valid value is between `0` and `100`.
* `tunnel1_rekey_margin_time_seconds` - (Optional, Default `540`) The margin time, in seconds, before the phase 2 lifetime expires, during which the AWS side of the first VPN connection performs an IKE rekey. The exact time of the rekey is randomly selected based on the value for `tunnel1_rekey_fuzz_percentage`. Valid value is between `60` and half of `tunnel1_phase2_lifetime_seconds`.
* `tunnel2_rekey_margin_time_seconds` - (Optional, Default `540`) The margin time, in seconds, before the phase 2 lifetime expires, during which the AWS side of the second VPN connection performs an IKE rekey. The exact time of the rekey is randomly selected based on the value for `tunnel2_rekey_fuzz_percentage`. Valid value is between `60` and half of `tunnel2_phase2_lifetime_seconds`.
* `tunnel1_replay_window_size` - (Optional, Default `1024`) The number of packets in an IKE replay window for the first VPN tunnel. Valid value is between `64` and `2048`.
* `tunnel2_replay_window_size` - (Optional, Default `1024`) The number of packets in an IKE replay window for the second VPN tunnel. Valid value is between `64` and `2048`.
* `tunnel1_startup_action` - (Optional, Default `add`) The action to take when the establishing the tunnel for the first VPN connection. By default, your customer gateway device must initiate the IKE negotiation and bring up the tunnel. Specify start for AWS to initiate the IKE negotiation. Valid values are `add | start`.
* `tunnel2_startup_action` - (Optional, Default `add`) The action to take when the establishing the tunnel for the second VPN connection. By default, your customer gateway device must initiate the IKE negotiation and bring up the tunnel. Specify start for AWS to initiate the IKE negotiation. Valid values are `add | start`.

### Log Options

The `tunnel1_log_options` and `tunnel2_log_options` block supports the following arguments:

* `cloudwatch_log_options` - (Optional) Options for sending VPN tunnel logs to CloudWatch. See [CloudWatch Log Options](#cloudwatch-log-options) below for more details.

### CloudWatch Log Options

The `cloudwatch_log_options` blocks supports the following arguments:

* `log_enabled` - (Optional) Enable or disable VPN tunnel logging feature. The default is `false`.
* `log_group_arn` - (Optional) The Amazon Resource Name (ARN) of the CloudWatch log group to send logs to.
* `log_output_format` - (Optional) Set log format. Default format is json. Possible values are: `json` and `text`. The default is `json`.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - Amazon Resource Name (ARN) of the VPN Connection.
* `id` - The amazon-assigned ID of the VPN connection.
* `core_network_arn` - The ARN of the core network.
* `core_network_attachment_arn` - The ARN of the core network attachment.
* `customer_gateway_configuration` - The configuration information for the VPN connection's customer gateway (in the native XML format).
* `customer_gateway_id` - The ID of the customer gateway to which the connection is attached.
* `routes` - The static routes associated with the VPN connection. Detailed below.
* `static_routes_only` - Whether the VPN connection uses static routes exclusively.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).
* `transit_gateway_attachment_id` - When associated with an EC2 Transit Gateway (`transit_gateway_id` argument), the attachment ID. See also the [`aws_ec2_tag` resource](/docs/providers/aws/r/ec2_tag.html) for tagging the EC2 Transit Gateway VPN Attachment.
* `tunnel1_address` - The public IP address of the first VPN tunnel.
* `tunnel1_cgw_inside_address` - The RFC 6890 link-local address of the first VPN tunnel (Customer Gateway Side).
* `tunnel1_vgw_inside_address` - The RFC 6890 link-local address of the first VPN tunnel (VPN Gateway Side).
* `tunnel1_preshared_key` - The preshared key of the first VPN tunnel.
* `tunnel1_bgp_asn` - The bgp asn number of the first VPN tunnel.
* `tunnel1_bgp_holdtime` - The bgp holdtime of the first VPN tunnel.
* `tunnel2_address` - The public IP address of the second VPN tunnel.
* `tunnel2_cgw_inside_address` - The RFC 6890 link-local address of the second VPN tunnel (Customer Gateway Side).
* `tunnel2_vgw_inside_address` - The RFC 6890 link-local address of the second VPN tunnel (VPN Gateway Side).
* `tunnel2_preshared_key` - The preshared key of the second VPN tunnel.
* `tunnel2_bgp_asn` - The bgp asn number of the second VPN tunnel.
* `tunnel2_bgp_holdtime` - The bgp holdtime of the second VPN tunnel.
* `vgw_telemetry` - Telemetry for the VPN tunnels. Detailed below.
* `vpn_gateway_id` - The ID of the virtual private gateway to which the connection is attached.

### routes

* `destination_cidr_block` - The CIDR block associated with the local subnet of the customer data center.
* `source` - Indicates how the routes were provided.
* `state` - The current state of the static route.

### vgw_telemetry

* `accepted_route_count` - The number of accepted routes.
* `certificate_arn` - The Amazon Resource Name (ARN) of the VPN tunnel endpoint certificate.
* `last_status_change` - The date and time of the last change in status.
* `outside_ip_address` - The Internet-routable IP address of the virtual private gateway's outside interface.
* `status` - The status of the VPN tunnel.
* `status_message` - If an error occurs, a description of the error.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import VPN Connections using the VPN connection `id`. For example:

```terraform
import {
  to = aws_vpn_connection.testvpnconnection
  id = "vpn-40f41529"
}
```

Using `terraform import`, import VPN Connections using the VPN connection `id`. For example:

```console
% terraform import aws_vpn_connection.testvpnconnection vpn-40f41529
```
