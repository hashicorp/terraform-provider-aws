---
subcategory: "VPC (Virtual Private Cloud)"
layout: "aws"
page_title: "AWS: aws_network_interface"
description: |-
  Get information on a Network Interface resource.
---

# Data Source: aws_network_interface

Use this data source to get information about a Network Interface.

## Example Usage

```terraform
data "aws_network_interface" "bar" {
  id = "eni-01234567"
}
```

## Argument Reference

This data source supports the following arguments:

* `account` - (Optional) AWS account ID where this data source will be read. Defaults to the account of the provider's credentials. See the [Per-Resource Account guide](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/guides/per-resource-account).
* `filter` - (Optional) One or more name/value pairs to filter off of. There are several valid keys, for a full reference, check out [describe-network-interfaces](https://docs.aws.amazon.com/cli/latest/reference/ec2/describe-network-interfaces.html) in the AWS CLI reference.
* `id` - (Optional) Identifier for the network interface.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

### `filter` Block

* `name` - (Required) Name of the field to filter by, as defined by [the underlying AWS API](https://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_DescribeNetworkInterfaces.html).
* `values` - (Required) Set of values that are accepted for the given field.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `arn` - ARN of the network interface.
* `association` - Association information for an Elastic IP address (IPv4) associated with the network interface. See [association](#association-block) below.
* `attachment` - Attachment of the ENI. See [attachment](#attachment-block) below.
* `availability_zone` - Availability Zone.
* `description` - Description of the network interface.
* `ena_srd_specification` - ENA Express configuration for the network interface. See [ena_srd_specification](#ena_srd_specification-block) below.
* `interface_type` - Type of interface.
* `ipv6_addresses` - List of IPv6 addresses to assign to the ENI.
* `mac_address` - MAC address.
* `outpost_arn` - ARN of the Outpost.
* `owner_id` - AWS account ID of the owner of the network interface.
* `private_dns_name` - Private DNS name.
* `private_ip` - Private IPv4 address of the network interface within the subnet.
* `private_ips` - Private IPv4 addresses associated with the network interface.
* `requester_id` - ID of the entity that launched the instance on your behalf.
* `security_groups` - List of security groups for the network interface.
* `subnet_id` - ID of the subnet.
* `tags` - Any tags assigned to the network interface.
* `vpc_id` - ID of the VPC.

### `association` Block

* `allocation_id` - Allocation ID.
* `association_id` - Association ID.
* `carrier_ip` - Carrier IP address associated with the network interface. This attribute is only set when the network interface is in a subnet which is associated with a Wavelength Zone.
* `customer_owned_ip` - Customer-owned IP address.
* `ip_owner_id` - ID of the Elastic IP address owner.
* `public_dns_name` - Public DNS name.
* `public_ip` - Address of the Elastic IP address bound to the network interface.

### `attachment` Block

* `attachment_id` - ID of the network interface attachment.
* `device_index` - Device index of the network interface attachment on the instance.
* `instance_id` - ID of the instance.
* `instance_owner_id` - AWS account ID of the owner of the instance.
* `network_card_index` - Index of the network card.

### `ena_srd_specification` Block

* `ena_srd_enabled` - Whether ENA Express is enabled for the network interface.
* `ena_srd_udp_specification` - ENA Express UDP configuration. See below.

#### `ena_srd_udp_specification` Block

* `ena_srd_udp_enabled` - Whether UDP traffic uses ENA Express.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

- `read` - (Default `20m`)
