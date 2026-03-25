---
subcategory: "WorkMail"
layout: "aws"
page_title: "AWS: aws_workmail_organization"
description: |-
  Lists WorkMail Organization resources.
---

# List Resource: aws_workmail_organization

Lists WorkMail Organization resources.

## Example Usage

```terraform
list "aws_workmail_organization" "example" {
  provider = aws
}
```

## Argument Reference

This list resource supports the following arguments:

* `region` - (Optional) Region to query. Defaults to provider region.
