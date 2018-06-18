---
layout: "aws"
page_title: "AWS: aws_vpc_dhcp_options"
sidebar_current: "docs-aws-datasource-vpc-dhcp-options"
description: |-
  Retrieve information about an EC2 DHCP Options configuration
---

# Data Source: aws_vpc_dhcp_options

Retrieve information about an EC2 DHCP Options configuration.

## Example Usage

```hcl
data "aws_vpc_dhcp_options" "test" {
  dhcp_options_id = "dopts-12345678"
}
```

## Argument Reference

* `dhcp_options_id` - (Required) The EC2 DHCP Options ID.

## Attributes Reference

* `domain_name` - The suffix domain name to used when resolving non Fully Qualified Domain Names. e.g. the `search` value in the `/etc/resolv.conf` file.
* `domain_name_servers` - List of name servers.
* `netbios_name_servers` - List of NETBIOS name servers.
* `netbios_node_type` - The NetBIOS node type (1, 2, 4, or 8). For more information about these node types, see [RFC 2132](http://www.ietf.org/rfc/rfc2132.txt).
* `ntp_servers` - List of NTP servers.
* `tags` - A mapping of tags assigned to the resource.
