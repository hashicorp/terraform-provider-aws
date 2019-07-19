---
layout: "aws"
page_title: "AWS: aws_organizations_organization"
sidebar_current: "docs-aws-datasource-organizations-organization"
description: |-
  Provides details about an organization.
---

# Data Source: aws_organizations_organization

Retrieves information about the organization that the user's account belongs to.

## Example Usage:

```hcl
data "aws_organizations_organization" "org" {}
```

## Argument Reference

N/A

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - ARN of the organization
* `id` - Identifier of the organization
* `master_account_arn` - ARN of the master account
* `master_account_email` - Email address of the master account
* `master_account_id` - Identifier of the master account
* `available_policy_types` - A list of policy type objects that are enabed for this organization

The policy type object contains the following attributes:

* `type` - The name of the policy type
* `status` - The status of the policy type as it relates to the associated root.