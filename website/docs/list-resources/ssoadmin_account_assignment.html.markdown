---
subcategory: "SSO Admin"
layout: "aws"
page_title: "AWS: aws_ssoadmin_account_assignment"
description: |-
  Lists SSO Admin Account Assignment resources.
---

# List Resource: aws_ssoadmin_account_assignment

Lists SSO Admin Account Assignment resources.

## Example Usage

```terraform
list "aws_ssoadmin_account_assignment" "example" {
  provider = aws

  config {
    account_id         = data.aws_caller_identity.current.account_id
    instance_arn       = aws_ssoadmin_permission_set.example.instance_arn
    permission_set_arn = aws_ssoadmin_permission_set.example.arn
  }
}
```

## Argument Reference

This list resource supports the following arguments:

* `account_id` - (Required) AWS account ID to list account assignments for.
* `instance_arn` - (Required) ARN of the IAM Identity Center instance.
* `permission_set_arn` - (Required) ARN of the permission set whose assignments should be listed.
* `region` - (Optional) Region to query. Defaults to provider region.
