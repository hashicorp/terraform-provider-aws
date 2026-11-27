---
subcategory: "EC2 (Elastic Compute Cloud)"
layout: "aws"
page_title: "AWS: aws_ec2_transit_gateway_policy_table_entry"
description: |-
  Lists EC2 (Elastic Compute Cloud) Transit Gateway Policy Table Entry resources.
---

# List Resource: aws_ec2_transit_gateway_policy_table_entry

Lists EC2 Transit Gateway Policy Table Entry resources.

## Example Usage

```terraform
list "aws_ec2_transit_gateway_policy_table_entry" "example" {
  provider = aws

  config {
    transit_gateway_policy_table_id = aws_ec2_transit_gateway_policy_table.example.id
  }
}
```

## Argument Reference

This list resource supports the following arguments:

* `account` - (Optional) AWS account ID to query. Defaults to the account of the provider's credentials. See the [Per-Resource Account guide](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/guides/per-resource-account).
* `region` - (Optional) Region to query. Defaults to provider region.
* `transit_gateway_policy_table_id` - (Required) ID of the transit gateway policy table.
