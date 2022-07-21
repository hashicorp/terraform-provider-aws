---
subcategory: "Organizations"
layout: "aws"
page_title: "AWS: aws_organizations_policy_attachments"
description: |-
  Get all policies that are attached to the specified target root, organizational unit (OU), or account.
---

# Data Source: aws_organizations_policy_attachments

Get all policies that are attached to the specified target root, organizational unit (OU), or account.

## Example Usage

```terraform
data "aws_organizations_organization" "org" {}

data "aws_organizations_policy_attachments" "attachments" {
  target_id = data.aws_organizations_organization.org.roots[0].id
  filter    = "SERVICE_CONTROL_POLICY"
}
```

## Argument Reference

* `target_id` - (Required) The unique identifier (ID) of the root, organizational unit, or account whose policies you want to list.
* `filter` - (Required) The type of policy that you want to include in the returned list. Valid values are `SERVICE_CONTROL_POLICY`, `TAG_POLICY`, `BACKUP_POLICY`, and `AISERVICES_OPT_OUT_POLICY`.

## Attributes Reference

* `policies` - List of child organizational units, which have the following attributes:
    * `arn` - Amazon Resource Name (ARN) of the policy
    * `aws_managed` - boolean value that indicates whether the specified policy is an AWS managed policy
    * `description` - description of the policy
    * `id` - unique identifier (ID) of the policy
    * `name` - friendly name of the policy
    * `type` - type of policy.
