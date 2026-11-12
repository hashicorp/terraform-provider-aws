---
subcategory: "SSO Admin"
layout: "aws"
page_title: "AWS: aws_ssoadmin_managed_policy_attachment"
description: |-
  Lists SSO Admin Managed Policy Attachment resources.
---

# List Resource: aws_ssoadmin_managed_policy_attachment

Lists SSO Admin Managed Policy Attachment resources.

## Example Usage

```terraform
list "aws_ssoadmin_managed_policy_attachment" "example" {
  provider = aws

  config {
    instance_arn       = aws_ssoadmin_permission_set.example.instance_arn
    permission_set_arn = aws_ssoadmin_permission_set.example.arn
  }
}
```

## Argument Reference

This list resource supports the following arguments:

* `instance_arn` - (Required) ARN of the IAM Identity Center instance.
* `permission_set_arn` - (Required) ARN of the permission set whose managed policy attachments should be listed.
* `region` - (Optional) Region to query. Defaults to provider region.
