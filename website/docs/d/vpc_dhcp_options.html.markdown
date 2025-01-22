---
subcategory: "VPC (Virtual Private Cloud)"
layout: "aws"
page_title: "AWS: aws_vpc_dhcp_options"
description: |-
  Retrieve information about an EC2 DHCP Options configuration
---

# Data Source: aws_vpc_dhcp_options

Retrieve information about an EC2 DHCP Options configuration.

## Example Usage

### Lookup by DHCP Options ID

```terraform
data "aws_vpc_dhcp_options" "example" {
  dhcp_options_id = "dopts-12345678"
}
```

### Lookup by Filter

```terraform
data "aws_vpc_dhcp_options" "example" {
  filter {
    name   = "key"
    values = ["domain-name"]
  }

  filter {
    name   = "value"
    values = ["example.com"]
  }
}
```

## Argument Reference

* `dhcp_options_id` - (Optional) EC2 DHCP Options ID.
* `filter` - (Optional) List of custom filters as described below.

### filter

For more information about filtering, see the [EC2 API documentation](https://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_DescribeDhcpOptions.html).

* `name` - (Required) Name of the field to filter.
* `values` - (Required) Set of values for filtering.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `arn` - ARN of the DHCP Options Set.
* `dhcp_options_id` - EC2 DHCP Options ID
* `domain_name` - Suffix domain name to used when resolving non Fully Qualified Domain NamesE.g., the `search` value in the `/etc/resolv.conf` file.
* `domain_name_servers` - List of name servers.
* `id` - EC2 DHCP Options ID
* `ipv6_address_preferred_lease_time` - How frequently, in seconds, a running instance with an IPv6 assigned to it goes through DHCPv6 lease renewal.
* `netbios_name_servers` - List of NETBIOS name servers.
* `netbios_node_type` - NetBIOS node type (1, 2, 4, or 8). For more information about these node types, see [RFC 2132](http://www.ietf.org/rfc/rfc2132.txt).
* `ntp_servers` - List of NTP servers.
* `tags` - Map of tags assigned to the resource.
* `owner_id` - ID of the AWS account that owns the DHCP options set.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

- `read` - (Default `20m`)
