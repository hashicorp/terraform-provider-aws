---
subcategory: "Lambda Core"
layout: "aws"
page_title: "AWS: aws_lambdacore_network_connector"
description: |-
  Lists Lambda Core Network Connector resources.
---

# List Resource: aws_lambdacore_network_connector

Lists Lambda Core Network Connector resources.

## Example Usage

```terraform
list "aws_lambdacore_network_connector" "example" {
  provider = aws
}
```

## Argument Reference

This list resource supports the following arguments:

* `region` - (Optional) Region to query. Defaults to provider region.
