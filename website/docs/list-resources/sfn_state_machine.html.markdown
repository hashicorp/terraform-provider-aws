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

* `account` - (Optional) AWS account ID to query. Defaults to the account of the provider's credentials. See the [Per-Resource Account guide](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/guides/per-resource-account).
* `region` - (Optional) Region to query. Defaults to provider region.
