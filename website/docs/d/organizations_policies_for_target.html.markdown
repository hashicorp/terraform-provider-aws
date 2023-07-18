---
subcategory: "Organizations"
layout: "aws"
page_title: "AWS: aws_organizations_policies_for_target"
description: |-
  Terraform data source for managing an AWS Organizations Policies For Target.
---

# Data Source: aws_organizations_policies_for_target

Terraform data source for managing an AWS Organizations Policies For Target.

## Example Usage

### Basic Usage

```terraform
data "aws_organizations_organization" "example" {}

data "aws_organizations_policies_for_target" "example" {
  target_id = data.aws_organizations_organization.example.roots[0].id
  filter    = "SERVICE_CONTROL_POLICY"
}

data "aws_organizations_policy" "example" {
  for_each  = toset(data.aws_organizations_policies_for_target.example.ids)
  policy_id = each.value
}
```

## Argument Reference

The following arguments are required:

* `target_id` - (Required) The root (string that begins with "r-" followed by 4-32 lowercase letters or digits), account (12 digit string), or Organizational Unit (string starting with "ou-" followed by 4-32 lowercase letters or digits. This string is followed by a second "-" dash and from 8-32 additional lowercase letters or digits.)
* `filter` - (Required) Must supply one of the 4 different policy filters for a target (SERVICE_CONTROL_POLICY | TAG_POLICY | BACKUP_POLICY | AISERVICES_OPT_OUT_POLICY)

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `ids` - List of all the policy ids found.
