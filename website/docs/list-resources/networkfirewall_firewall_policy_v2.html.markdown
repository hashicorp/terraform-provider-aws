---
subcategory: "Network Firewall"
layout: "aws"
page_title: "AWS: aws_networkfirewall_firewall_policy_v2"
description: |-
  Lists Network Firewall Firewall Policy V2 resources.
---

# List Resource: aws_networkfirewall_firewall_policy_v2

Lists Network Firewall Firewall Policy V2 resources.

## Example Usage

```terraform
list "aws_networkfirewall_firewall_policy_v2" "example" {
  provider = aws
}
```

## Argument Reference

This list resource supports the following arguments:

* `region` - (Optional) Region to query. Defaults to provider region.
