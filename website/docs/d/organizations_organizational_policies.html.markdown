---
subcategory: "Organizations"
layout: "aws"
page_title: "AWS: aws_organizations_organizational_policies"
description: |-
  Terraform data source for managing an AWS Organizations Organization Policies.
---

# Data Source: aws_organizations_organizational_policies

Get all directly attached policy summaries for target root, organizational unit (OU), or account.

## Example Usage

```terraform
data "aws_organizations_organization" "org" {}

data "aws_organizations_organizational_policies" "policies" {
    target_id = data.aws_organizations_organization.org.roots[0].id
}
```

## Argument Reference

The following arguments are required:

* `target_id` - (Required) The root (string that begins with "r-" followed by 4-32 lowercase letters or digits), account (12 digit string), or Organizational Unit (string starting with "ou-" followed by 4-32 lowercase letters or digits. This string is followed by a second "-" dash and from 8-32 additional lowercase letters or digits.)

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `policies` - List of child accounts, which have the following attributes:
    * `arn` - The Amazon Resource Name (ARN) of the account.
      * `aws_managed` - Indicates if a policy is AWS managed.
      * `description` - Description of the policy.
      * `id` - The unique identifier (ID) of the policy.
      * `name` - The friendly name of the policy.
      * `type` - The type of policy.