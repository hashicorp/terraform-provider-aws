---
subcategory: "Lightsail"
layout: "aws"
page_title: "AWS: aws_lightsail_instance_public_ports"
description: |-
  Provides an Lightsail Instance
---

# Resource: aws_lightsail_instance_public_ports

Opens ports for a specific Amazon Lightsail instance, and specifies the IP addresses allowed to connect to the instance through the ports, and the protocol.

-> See [What is Amazon Lightsail?](https://lightsail.aws.amazon.com/ls/docs/getting-started/article/what-is-amazon-lightsail) for more information.

~> **Note:** Lightsail is currently only supported in a limited number of AWS Regions, please see ["Regions and Availability Zones in Amazon Lightsail"](https://lightsail.aws.amazon.com/ls/docs/overview/article/understanding-regions-and-availability-zones-in-amazon-lightsail) for more details.

## Example Usage

```terraform
resource "aws_lightsail_instance" "test" {
  name              = "yak_sail"
  availability_zone = data.aws_availability_zones.available.names[0]
  blueprint_id      = "amazon_linux_2"
  bundle_id         = "nano_3_0"
}

resource "aws_lightsail_instance_public_ports" "test" {
  instance_name = aws_lightsail_instance.test.name

  port_info {
    protocol  = "tcp"
    from_port = 80
    to_port   = 80
  }
}
```

## Argument Reference

The following arguments are required:

* `instance_name` - (Required) Name of the instance for which to open ports.
* `port_info` - (Required) Descriptor of the ports to open for the specified instance. AWS closes all currently open ports that are not included in this argument. See [`port_info` Block](#port_info-block) for details.

### `port_info` Block

The `port_info` configuration block supports the following arguments:

* `from_port` - (Required) First port in a range of open ports on an instance. See [PortInfo](https://docs.aws.amazon.com/lightsail/2016-11-28/api-reference/API_PortInfo.html) for details.
* `protocol` - (Required) IP protocol name. Valid values: `tcp`, `all`, `udp`, `icmp`, `icmpv6`. See [PortInfo](https://docs.aws.amazon.com/lightsail/2016-11-28/api-reference/API_PortInfo.html) for details.
* `to_port` - (Required) Last port in a range of open ports on an instance. See [PortInfo](https://docs.aws.amazon.com/lightsail/2016-11-28/api-reference/API_PortInfo.html) for details.
* `cidr_list_aliases` - (Optional) Set of CIDR aliases that define access for a preconfigured range of IP addresses.
* `cidrs` - (Optional) Set of IPv4 addresses or ranges of IPv4 addresses (in CIDR notation) that are allowed to connect to an instance through the ports, and the protocol.
* `ipv6_cidrs` - (Optional) Set of IPv6 addresses or ranges of IPv6 addresses (in CIDR notation) that are allowed to connect to an instance through the ports, and the protocol.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - ID of the resource.
