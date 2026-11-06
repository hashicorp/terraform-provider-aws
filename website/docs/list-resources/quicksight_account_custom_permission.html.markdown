---
subcategory: "QuickSight"
layout: "aws"
page_title: "AWS: aws_quicksight_account_custom_permission"
description: |-
  Lists QuickSight Account Custom Permission resources.
---

# List Resource: aws_quicksight_account_custom_permission

Lists QuickSight Account Custom Permission resources.

## Example Usage

```terraform
list "aws_quicksight_account_custom_permission" "example" {
  provider = aws

  include_resource = true
}
```

## Argument Reference

This list resource supports the following arguments:

* `region` - (Optional) Region to query. Defaults to provider region.
