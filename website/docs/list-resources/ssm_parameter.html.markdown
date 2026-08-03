---
subcategory: "SSM (Systems Manager)"
layout: "aws"
page_title: "AWS: aws_ssm_parameter"
description: |-
  Lists SSM Parameter resources.
---

# List Resource: aws_ssm_parameter

Lists SSM Parameter resources.

## Example Usage

```terraform
list "aws_ssm_parameter" "example" {
  provider = aws
}
```

## Argument Reference

This list resource supports the following arguments:

* `region` - (Optional) Region to query. Defaults to provider region.
