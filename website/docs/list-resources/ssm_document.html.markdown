---
subcategory: "SSM (Systems Manager)"
layout: "aws"
page_title: "AWS: aws_ssm_document"
description: |-
  Lists SSM (Systems Manager) Document resources.
---

# List Resource: aws_ssm_document

Lists SSM (Systems Manager) Document resources.

## Example Usage

```terraform
list "aws_ssm_document" "example" {
  provider = aws
}
```

## Argument Reference

This list resource supports the following arguments:

* `region` - (Optional) Region to query. Defaults to provider region.
