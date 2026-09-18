---
subcategory: "SES Mail Manager"
layout: "aws"
page_title: "AWS: aws_mailmanager_archive"
description: |-
  Lists SES Mail Manager Archive resources.
---

# List Resource: aws_mailmanager_archive

Lists SES Mail Manager Archive resources.

## Example Usage

```terraform
list "aws_mailmanager_archive" "example" {
  provider = aws
}
```

## Argument Reference

This list resource supports the following arguments:

* `region` - (Optional) Region to query. Defaults to provider region.
