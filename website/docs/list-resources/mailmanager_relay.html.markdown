---
subcategory: "SES Mail Manager"
layout: "aws"
page_title: "AWS: aws_mailmanager_relay"
description: |-
  Lists SES Mail Manager Relay resources.
---

# List Resource: aws_mailmanager_relay

Lists SES Mail Manager Relay resources.

## Example Usage

```terraform
list "aws_mailmanager_relay" "example" {
  provider = aws
}
```

## Argument Reference

This list resource supports the following arguments:

* `region` - (Optional) Region to query. Defaults to provider region.
