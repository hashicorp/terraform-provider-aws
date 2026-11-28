---
subcategory: "EC2 (Elastic Compute Cloud)"
layout: "aws"
page_title: "AWS: aws_ec2_secondary_network"
description: |-
  Lists EC2 (Elastic Compute Cloud) Secondary Network resources.
---

# List Resource: aws_ec2_secondary_network

Lists EC2 Secondary Network resources.

## Example Usage

```terraform
list "aws_ec2_secondary_network" "example" {
  provider = aws
}
```

## Argument Reference

This list resource supports the following arguments:

* `account` - (Optional) AWS account ID to query. Defaults to the account of the provider's credentials. See the [Per-Resource Account guide](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/guides/per-resource-account).
* `region` - (Optional) Region to query. Defaults to provider region.
