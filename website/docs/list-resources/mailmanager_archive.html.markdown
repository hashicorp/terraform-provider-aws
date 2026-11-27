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

* `account` - (Optional) AWS account ID to query. Defaults to the account of the provider's credentials. See the [Per-Resource Account guide](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/guides/per-resource-account).
* `region` - (Optional) Region to query. Defaults to provider region.
