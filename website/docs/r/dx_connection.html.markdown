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

```terraform
resource "aws_dx_connection" "hoge" {
  name      = "tf-dx-connection"
  bandwidth = "1Gbps"
  location  = "EqDC2"
}
```

## Argument Reference

The following arguments are supported:

* `bandwidth` - (Required) The bandwidth of the connection. Valid values for dedicated connections: 1Gbps, 10Gbps. Valid values for hosted connections: 50Mbps, 100Mbps, 200Mbps, 300Mbps, 400Mbps, 500Mbps, 1Gbps, 2Gbps, 5Gbps, 10Gbps and 100Gbps. Case sensitive.
* `location` - (Required) The AWS Direct Connect location where the connection is located. See [DescribeLocations](https://docs.aws.amazon.com/directconnect/latest/APIReference/API_DescribeLocations.html) for the list of AWS Direct Connect locations. Use `locationCode`.
* `name` - (Required) The name of the connection.
* `provider_name` - (Optional) The name of the service provider associated with the connection.
* `tags` - (Optional) A map of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - The ARN of the connection.
* `aws_device` - The Direct Connect endpoint on which the physical connection terminates.
* `has_logical_redundancy` - Indicates whether the connection supports a secondary BGP peer in the same address family (IPv4/IPv6).
* `id` - The ID of the connection.
* `jumbo_frame_capable` - Boolean value representing if jumbo frames have been enabled for this connection.
* `owner_account_id` - The ID of the AWS account that owns the connection.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block).

## Import

Direct Connect connections can be imported using the `connection id`, e.g.,

```
$ terraform import aws_dx_connection.test_connection dxcon-ffre0ec3
```
