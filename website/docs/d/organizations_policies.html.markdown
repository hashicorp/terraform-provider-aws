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
```

## Argument Reference

The following arguments are required:

* `filter` - (Required) The type of policies to be returned in the response. Valid values are `SERVICE_CONTROL_POLICY | TAG_POLICY | BACKUP_POLICY | AISERVICES_OPT_OUT_POLICY`

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `policies` - List of policies that have the following attributes:
    * `arn` - The Amazon Resource Name (ARN) of the account.
    * `aws_managed` - A boolean value that indicates whether the specified policy is an AWS managed policy. If true, then you can attach the policy to roots, OUs, or accounts, but you cannot edit it.
    * `description` - The description of the policy.
    * `id` - The unique identifier (ID) of the policy.
    * `name` - The friendly name of the policy.
    * `type` - The type of policy.
