---
subcategory: "SSM (Systems Manager)"
layout: "aws"
page_title: "AWS: aws_ssm_patch_group"
description: |-
  Lists SSM (Systems Manager) Patch Group resources.
---

# List Resource: aws_ssm_patch_group

Lists SSM (Systems Manager) Patch Group resources.

## Example Usage

```terraform
list "aws_ssm_patch_group" "example" {
  provider = aws
}
```

## Argument Reference

This list resource supports the following arguments:

* `region` - (Optional) Region to query. Defaults to provider region.
