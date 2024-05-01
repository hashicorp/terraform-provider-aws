---
subcategory: "EC2 (Elastic Compute Cloud)"
layout: "aws"
page_title: "AWS: aws_eip"
description: |-
    Provides details about a specific Elastic IP
---

# Data Source: aws_eip

`aws_eip` provides details about a specific Elastic IP.

## Example Usage

### Search By Allocation ID (VPC only)

```terraform
data "aws_eip" "by_allocation_id" {
  id = "eipalloc-12345678"
}
```

### Search By Filters (EC2-Classic or VPC)

```terraform
data "aws_eip" "by_filter" {
  filter {
    name   = "tag:Name"
    values = ["exampleNameTagValue"]
  }
}
```

### Search By Public IP (EC2-Classic or VPC)

```terraform
data "aws_eip" "by_public_ip" {
  public_ip = "1.2.3.4"
}
```

### Search By Tags (EC2-Classic or VPC)

```terraform
data "aws_eip" "by_tags" {
  tags = {
    Name = "exampleNameTagValue"
  }
}
```

## Argument Reference

The arguments of this data source act as filters for querying the available
Elastic IPs in the current region. The given filters must match exactly one
Elastic IP whose data will be exported as attributes.

* `filter` - (Optional) One or more name/value pairs to use as filters. There are several valid keys, for a full reference, check out the [EC2 API Reference](https://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_DescribeAddresses.html).
* `id` - (Optional) Allocation ID of the specific VPC EIP to retrieve. If a classic EIP is required, do NOT set `id`, only set `public_ip`
* `public_ip` - (Optional) Public IP of the specific EIP to retrieve.
* `tags` - (Optional) Map of tags, each pair of which must exactly match a pair on the desired Elastic IP

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `association_id` - ID representing the association of the address with an instance in a VPC.
* `carrier_ip` - Carrier IP address.
* `customer_owned_ip` - Customer Owned IP.
* `customer_owned_ipv4_pool` - The ID of a Customer Owned IP Pool. For more on customer owned IP addressed check out [Customer-owned IP addresses guide](https://docs.aws.amazon.com/outposts/latest/userguide/outposts-networking-components.html#ip-addressing)
* `domain` - Whether the address is for use in EC2-Classic (standard) or in a VPC (vpc).
* `id` - If VPC Elastic IP, the allocation identifier. If EC2-Classic Elastic IP, the public IP address.
* `instance_id` - ID of the instance that the address is associated with (if any).
* `network_interface_id` - The ID of the network interface.
* `network_interface_owner_id` - The ID of the AWS account that owns the network interface.
* `private_ip` - Private IP address associated with the Elastic IP address.
* `private_dns` - Private DNS associated with the Elastic IP address.
* `ptr_record` - The DNS pointer (PTR) record for the IP address.
* `public_ip` - Public IP address of Elastic IP.
* `public_dns` - Public DNS associated with the Elastic IP address.
* `public_ipv4_pool` - ID of an address pool.
* `tags` - Key-value map of tags associated with Elastic IP.

~> **Note:** The data source computes the `public_dns` and `private_dns` attributes according to the [VPC DNS Guide](https://docs.aws.amazon.com/vpc/latest/userguide/vpc-dns.html#vpc-dns-hostnames) as they are not available with the EC2 API.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

- `read` - (Default `20m`)
