---
subcategory: "FIS (Fault Injection Simulator)"
layout: "aws"
page_title: "AWS: aws_fis_safety_lever_state"
description: |-
  Lists the FIS (Fault Injection Simulator) safety lever state.
---

# List Resource: aws_fis_safety_lever_state

Lists the FIS (Fault Injection Simulator) safety lever state.

The safety lever is an account- and Region-level singleton, so a query returns at
most one result per queried Region.

## Example Usage

```terraform
list "aws_fis_safety_lever_state" "example" {
  provider = aws
}
```

## Argument Reference

This list resource supports the following arguments:

* `region` - (Optional) Region to query. Defaults to provider region.
