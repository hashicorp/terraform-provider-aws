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

* `account` - (Optional) AWS account ID to query. Defaults to the account of the provider's credentials. See the [Per-Resource Account guide](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/guides/per-resource-account).
* `region` - (Optional) Region to query. Defaults to provider region.
