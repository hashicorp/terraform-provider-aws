---
subcategory: "EC2"
layout: "aws"
page_title: "AWS: aws_ec2_client_vpn_endpoint"
description: |-
  Get information on an EC2 Client VPN Endpoint
---

# Data Source: aws_ec2_client_vpn_endpoint

Get information on an EC2 Client VPN Endpoint.

## Example Usage

### By Filter

```hcl
data "aws_ec2_client_vpn_endpoint" "example" {
  filter {
    name   = "tag:Name"
    values = ["ExampleVpn"]
  }
}
```

### By Identifier

```hcl
data "aws_ec2_client_vpn_endpoint" "example" {
  client_vpn_endpoint_id = "cvpn-endpoint-083cf50d6eb314f21"
}
```

## Argument Reference

The following arguments are supported:

* `client_vpn_endpoint_id` - (Optional) The ID of the Client VPN Endpoint.
* `filter` - (Optional) One or more configuration blocks containing name-values filters. Detailed below.
* `tags` - (Optional) Map of tags, each pair of which must exactly match a pair on the desired endpoint.

### filter

This block allows for complex filters. You can use one or more `filter` blocks.

The following arguments are required:

* `name` - (Required) The name of the field to filter by, as defined by [the underlying AWS API](https://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_DescribeClientVpnEndpoints.html).
* `values` - (Required) Set of values that are accepted for the given field. An endpoint will be selected if any one of the given values matches.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `arn` -  The ARN of the Client VPN Endpoint.
