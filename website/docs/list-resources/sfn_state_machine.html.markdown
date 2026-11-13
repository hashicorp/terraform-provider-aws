---
subcategory: "SFN (Step Functions)"
layout: "aws"
page_title: "AWS: aws_sfn_state_machine"
description: |-
  Lists SFN (Step Functions) State Machine resources.
---

# List Resource: aws_sfn_state_machine

Lists SFN (Step Functions) State Machine resources.

## Example Usage

```terraform
list "aws_sfn_state_machine" "example" {
  provider = aws
}
```

## Argument Reference

This list resource supports the following arguments:

* `region` - (Optional) Region to query. Defaults to provider region.
