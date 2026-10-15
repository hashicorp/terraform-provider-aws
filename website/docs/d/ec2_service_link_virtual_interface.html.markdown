---
subcategory: "Outposts (EC2)"
layout: "aws"
page_title: "AWS: aws_ec2_service_link_virtual_interface"
description: |-
    Provides details about an EC2 Service Link Virtual Interface
---

# Data Source: aws_ec2_service_link_virtual_interface

Provides details about an EC2 Service Link Virtual Interface. More information can be found in the [Outposts User Guide](https://docs.aws.amazon.com/outposts/latest/userguide/how-outposts-works.html#how-service-link).

## Example Usage

```terraform
data "aws_ec2_service_link_virtual_interface" "example" {
  id = "slvif-1234567890abcdef0"
}
```

## Argument Reference

The following arguments are optional:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `filter` - (Optional) One or more configuration blocks containing name-values filters. See the [EC2 API Reference](https://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_DescribeServiceLinkVirtualInterfaces.html) for supported filters. Detailed below.
* `id` - (Optional) Identifier of the EC2 Service Link Virtual Interface.

~> **NOTE:** At least one of `filter` or `id` must be specified.

### filter Argument Reference

The `filter` configuration block supports the following arguments:

* `name` - (Required) Name of the filter.
* `values` - (Required) List of one or more values for the filter.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `arn` - ARN of the Service Link Virtual Interface.
* `configuration_state` - Current state of the Service Link Virtual Interface.
* `local_address` - IPv4 address assigned to the local gateway virtual interface on the Outpost side.
* `outpost_arn` - Outpost ARN for the Service Link Virtual Interface.
* `outpost_id` - Outpost ID for the Service Link Virtual Interface.
* `outpost_lag_id` - Link aggregation group (LAG) ID for the Service Link Virtual Interface.
* `owner_id` - ID of the AWS account that owns the Service Link Virtual Interface.
* `peer_address` - IPv4 peer address for the Service Link Virtual Interface.
* `peer_bgp_asn` - BGP Autonomous System Number (ASN) of the peer.
* `tags` - Key-value map of resource tags.
* `vlan` - Virtual Local Area Network.
