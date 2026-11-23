---
subcategory: "Network Firewall"
layout: "aws"
page_title: "AWS: aws_networkfirewall_firewall"
description: |-
  Lists Network Firewall firewall resources.
---

# List Resource: aws_networkfirewall_firewall

Lists Network Firewall firewall resources.

## Example Usage

```terraform
list "aws_networkfirewall_firewall" "example" {
  provider = aws
}
```

## Argument Reference

This list resource supports the following arguments:

* `region` - (Optional) Region to query. Defaults to provider region.
