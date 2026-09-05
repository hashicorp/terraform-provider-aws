---
subcategory: "Direct Connect"
layout: "aws"
page_title: "AWS: aws_dx_connection"
description: |-
  Retrieve information about a Direct Connect Connection.
---

# Data Source: aws_dx_connection

Retrieve information about a Direct Connect Connection.

~> **Note:** This data source is different from the [`aws_dx_connections`](/docs/providers/aws/d/dx_connections.html) data source which retrieves information about all the AWS Direct Connect connections in the current AWS Region.

## Example Usage

```terraform
data "aws_dx_connection" "example" {
  name = "tf-dx-connection"
}
```

## Argument Reference

This data source supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `name` - (Required) Name of the connection to retrieve.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `arn` - ARN of the connection.
* `aws_device` - Direct Connect endpoint on which the physical connection terminates.
* `bandwidth` - Bandwidth of the connection.
* `id` - ID of the connection.
* `location` - AWS Direct Connect location where the connection is located.
* `owner_account_id` - ID of the AWS account that owns the connection.
* `partner_name` - The name of the AWS Direct Connect service provider associated with the connection.
* `prefix_pool_size_ipv4` - The total number of inbound IPv4 route prefixes that can be allocated across the virtual interfaces on the connection.
* `prefix_pool_size_ipv6` - The total number of inbound IPv6 route prefixes that can be allocated across the virtual interfaces on the connection.
* `prefix_pool_unallocated_count_ipv4` - The number of inbound IPv4 route prefixes in the connection prefix pool not yet allocated to a virtual interface.
* `prefix_pool_unallocated_count_ipv6` - The number of inbound IPv6 route prefixes in the connection prefix pool not yet allocated to a virtual interface.
* `provider_name` - Name of the service provider associated with the connection.
* `rate_limiter_status` - Rate limiter status for the connection. See [`rate_limiter_status` Block](#rate_limiter_status-block) below.
* `state` - State of the connection.
* `tags` - Map of tags for the resource.
* `vlan_id` - The VLAN ID.

### `rate_limiter_status` Block

`rate_limiter_status` exports the following attributes:

* `max_allowed` - Maximum number of rate limiters allowed on the connection.
* `in_use` - Number of rate limiters currently in use.
* `remaining` - Number of rate limiters remaining (available).
* `total_bandwidth` - Total bandwidth allocated across all rate limiters.
