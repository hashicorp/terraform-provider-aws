---
subcategory: "Account Access"
layout: "aws"
page_title: "AWS: aws_accountaccess_entitlement"
description: |-
  Lists Account Access Entitlement resources.
---

# List Resource: aws_accountaccess_entitlement

Lists Account Access Entitlement resources for an Application and target AWS account.

## Example Usage

```terraform
list "aws_accountaccess_entitlement" "example" {
  provider = aws

  config {
    account_id      = "123456789012"
    application_arn = "arn:aws:account-access:us-east-1:123456789012:application/aam-0123456789abcdef"
  }
}
```

## Argument Reference

This list resource supports the following arguments:

* `account_id` - (Required) AWS account ID to filter Entitlements by.
* `application_arn` - (Required) ARN of the parent Account Access Application to list Entitlements from.
* `region` - (Optional) Region to query. Defaults to provider region.
