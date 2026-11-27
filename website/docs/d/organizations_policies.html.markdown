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

* `account` - (Optional) AWS account ID where this data source will be read. Defaults to the account of the provider's credentials. See the [Per-Resource Account guide](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/guides/per-resource-account).
* `filter` - (Required) Type of policies to be returned in the response. Valid values are `AISERVICES_OPT_OUT_POLICY | BACKUP_POLICY | CHATBOT_POLICY | DECLARATIVE_POLICY_EC2 | RESOURCE_CONTROL_POLICY | SERVICE_CONTROL_POLICY | TAG_POLICY`

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `ids` - List of all the policy IDs found.
