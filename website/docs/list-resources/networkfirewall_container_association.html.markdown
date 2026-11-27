---
subcategory: "Network Firewall"
layout: "aws"
page_title: "AWS: aws_networkfirewall_container_association"
description: |-
  Lists Network Firewall Container Association resources.
---

# List Resource: aws_networkfirewall_container_association

Lists Network Firewall Container Association resources.

## Example Usage

```terraform
list "aws_networkfirewall_container_association" "example" {
  provider = aws
}
```

## Argument Reference

This list resource supports the following arguments:

* `account` - (Optional) AWS account ID to query. Defaults to the account of the provider's credentials. See the [Per-Resource Account guide](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/guides/per-resource-account).
* `region` - (Optional) Region to query. Defaults to provider region.
