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

* `region` - (Optional) Region to query. Defaults to provider region.
