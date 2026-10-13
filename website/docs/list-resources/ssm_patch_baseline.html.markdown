---
subcategory: "SSM (Systems Manager)"
layout: "aws"
page_title: "AWS: aws_ssm_patch_baseline"
description: |-
  Lists SSM (Systems Manager) Patch Baseline resources.
---

# List Resource: aws_ssm_patch_baseline

Lists SSM (Systems Manager) Patch Baseline resources.

## Example Usage

```terraform
list "aws_ssm_patch_baseline" "example" {
  provider = aws
}
```

## Argument Reference

This list resource supports the following arguments:

* `region` - (Optional) Region to query. Defaults to provider region.
