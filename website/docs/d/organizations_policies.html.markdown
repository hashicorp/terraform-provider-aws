---
subcategory: "Organizations"
layout: "aws"
page_title: "AWS: aws_organizations_policies"
description: |-
  Terraform data source for managing an AWS Organizations Policies.
---

# Data Source: aws_organizations_policies

Terraform data source for managing an AWS Organizations Policies.

## Example Usage

### Basic Usage

```terraform
data "aws_organizations_policies" "example" {
  filter = "SERVICE_CONTROL_POLICY"
}

data "aws_organizations_policy" "example" {
  for_each  = toset(data.aws_organizations_policies.example.ids)
  policy_id = each.value
}
```

## Argument Reference

The following arguments are required:

* `filter` - (Required) The type of policies to be returned in the response. Valid values are `SERVICE_CONTROL_POLICY | TAG_POLICY | BACKUP_POLICY | AISERVICES_OPT_OUT_POLICY`

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `ids` - List of all the policy ids found.
