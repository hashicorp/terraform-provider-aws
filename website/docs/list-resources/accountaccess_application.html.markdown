---
subcategory: "Account Access"
layout: "aws"
page_title: "AWS: aws_accountaccess_application"
description: |-
  Lists Account Access Application resources.
---

# List Resource: aws_accountaccess_application

Lists Account Access Application resources.

## Example Usage

```terraform
list "aws_accountaccess_application" "example" {
  provider = aws
}
```

## Argument Reference

This list resource supports the following arguments:

* `region` - (Optional) Region to query. Defaults to provider region.
