---
subcategory: "EC2 (Elastic Compute Cloud)"
layout: "aws"
page_title: "AWS: aws_ec2_transit_gateway_metering_policy"
description: |-
  Lists EC2 (Elastic Compute Cloud) Transit Gateway Metering Policy resources.
---

# List Resource: aws_ec2_transit_gateway_metering_policy

Lists EC2 Transit Gateway Metering Policy resources.

## Example Usage

```terraform
list "aws_ec2_transit_gateway_metering_policy" "example" {
  provider = aws
}
```

## Argument Reference

This list resource supports the following arguments:

* `account` - (Optional) AWS account ID to query. Defaults to the account of the provider's credentials. See the [Per-Resource Account guide](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/guides/per-resource-account).
* `region` - (Optional) Region to query. Defaults to provider region.
