---
subcategory: "VPC (Virtual Private Cloud)"
layout: "aws"
page_title: "AWS: aws_network_interface"
description: |-
  Provides an Elastic network interface (ENI) resource.
---

# Resource: aws_network_interface

Provides an Elastic network interface (ENI) resource.

## Example Usage

```terraform
resource "aws_network_interface" "test" {
  subnet_id       = aws_subnet.public_a.id
  private_ips     = ["10.0.0.50"]
  security_groups = [aws_security_group.web.id]

  attachment {
    instance     = aws_instance.test.id
    device_index = 1
  }
}
```

### Example of Managing Multiple IPs on a Network Interface

By default, private IPs are managed through the `private_ips` and `private_ips_count` arguments which manage IPs as a set of IPs that are configured without regard to order. For a new network interface, the same primary IP address is consistently selected from a given set of addresses, regardless of the order provided. However, modifications of the set of addresses of an existing interface will not alter the current primary IP address unless it has been removed from the set.

In order to manage the private IPs as a sequentially ordered list, configure `private_ip_list_enabled` to `true` and use `private_ip_list` to manage the IPs. This will disable the `private_ips` and `private_ips_count` settings, which must be removed from the config file but are still exported. Note that changing the first address of `private_ip_list`, which is the primary, always requires a new interface.

If you are managing a specific set or list of IPs, instead of just using `private_ips_count`, this is a potential workflow for also leveraging `private_ips_count` to have AWS automatically assign additional IP addresses:

1. Comment out `private_ips`, `private_ip_list`, `private_ip_list_enabled` in your configuration
2. Set the desired `private_ips_count` (count of the number of secondaries, the primary is not included)
3. Apply to assign the extra IPs
4. Remove `private_ips_count` and restore your settings from the first step
5. Add the new IPs to your current settings
6. Apply again to update the stored state

This process can also be used to remove IP addresses in addition to the option of manually removing them. Adding IP addresses in a manually is more difficult because it requires knowledge of which addresses are available.

## Argument Reference

The following arguments are required:

* `subnet_id` - (Required) Subnet ID to create the ENI in.

The following arguments are optional:

* `attachment` - (Optional) Configuration block to define the attachment of the ENI. See [Attachment](#attachment) below for more details!
* `description` - (Optional) Description for the network interface.
* `interface_type` - (Optional) Type of network interface to create. Set to `efa` for Elastic Fabric Adapter. Changing `interface_type` will cause the resource to be destroyed and re-created.
* `ipv4_prefix_count` - (Optional) Number of IPv4 prefixes that AWS automatically assigns to the network interface.
* `ipv4_prefixes` - (Optional) One or more IPv4 prefixes assigned to the network interface.
* `ipv6_address_count` - (Optional) Number of IPv6 addresses to assign to a network interface. You can't use this option if specifying specific `ipv6_addresses`. If your subnet has the AssignIpv6AddressOnCreation attribute set to `true`, you can specify `0` to override this setting.
* `ipv6_address_list_enable` - (Optional) Whether `ipv6_addreses_list` is allowed and controls the IPs to assign to the ENI and `ipv6_addresses` and `ipv6_addresses_count` become read-only. Default false.
* `ipv6_address_list` - (Optional) List of private IPs to assign to the ENI in sequential order.
* `ipv6_addresses` - (Optional) One or more specific IPv6 addresses from the IPv6 CIDR block range of your subnet. Addresses are assigned without regard to order. You can't use this option if you're specifying `ipv6_address_count`.
* `ipv6_prefix_count` - (Optional) Number of IPv6 prefixes that AWS automatically assigns to the network interface.
* `ipv6_prefixes` - (Optional) One or more IPv6 prefixes assigned to the network interface.
* `private_ip_list` - (Optional) List of private IPs to assign to the ENI in sequential order. Requires setting `private_ip_list_enabled` to `true`.
* `private_ip_list_enabled` - (Optional) Whether `private_ip_list` is allowed and controls the IPs to assign to the ENI and `private_ips` and `private_ips_count` become read-only. Default false.
* `private_ips` - (Optional) List of private IPs to assign to the ENI without regard to order.
* `private_ips_count` - (Optional) Number of secondary private IPs to assign to the ENI. The total number of private IPs will be 1 + `private_ips_count`, as a primary private IP will be assiged to an ENI by default.
* `security_groups` - (Optional) List of security group IDs to assign to the ENI.
* `source_dest_check` - (Optional) Whether to enable source destination checking for the ENI. Default true.
* `tags` - (Optional) Map of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

### Attachment

The `attachment` block supports the following:

* `instance` - (Required) ID of the instance to attach to.
* `device_index` - (Required) Integer to define the devices index.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - ARN of the network interface.
* `id` - ID of the network interface.
* `mac_address` - MAC address of the network interface.
* `owner_id` - AWS account ID of the owner of the network interface.
* `private_dns_name` - Private DNS name of the network interface (IPv4).
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block).

## Import

Network Interfaces can be imported using the `id`, e.g.,

```
$ terraform import aws_network_interface.test eni-e5aa89a3
```
