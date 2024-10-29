---
subcategory: "Organizations"
layout: "aws"
page_title: "AWS: aws_organizations_policy"
description: |-
  Terraform data source for managing an AWS Organizations Policy.
---

# Data Source: aws_organizations_policy

Terraform data source for managing an AWS Organizations Policy.

## Example Usage

### Basic Usage

```terraform
data "aws_organizations_organization" "current" {}

data "aws_organizations_policies_for_target" "current" {
  target_id = data.aws_organizations_organization.current.roots[0].id
  filter    = "SERVICE_CONTROL_POLICY"
}
data "aws_organizations_policy" "test" {
  policy_id = data.aws_organizations_policies_for_target.current.policies[0].id
}
```

## Argument Reference

The following arguments are required:

* `policy_id` - (Required) The unique identifier (ID) of the policy that you want more details on. Policy id starts with a "p-" followed by 8-28 lowercase or uppercase letters, digits, and underscores.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `arn` - The Amazon Resource Name of the policy.
* `aws_managed` - Indicates if a policy is an AWS managed policy.
* `content` - The text content of the policy.
* `description` - The description of the policy.
* `name` - The friendly name of the policy.
* `type` - The type of policy values can be `SERVICE_CONTROL_POLICY | TAG_POLICY | BACKUP_POLICY | AISERVICES_OPT_OUT_POLICY`
