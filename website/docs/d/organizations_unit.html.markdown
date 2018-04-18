---
layout: "aws"
page_title: "AWS: aws_organizations_unit"
sidebar_current: "docs-aws-datasource-organizations-unit"
description: |-
    Provides details about an organizational unit
---

# Data Source: aws_organizations_unit

`aws_organizations_unit` provides details about an [organizational unit](https://docs.aws.amazon.com/organizations/latest/userguide/orgs_manage_ous.html).

~> **Note:** Only supports root organizational units at the moment. Also, must be retrieved from the organization's master account.

Will give an error if Organizations aren't enabled - see `aws_organizations_organization`.

## Example Usage

```hcl
data "aws_organizations_unit" "root" {
  root = true
}
```

## Argument Reference

* `root` - (Optional) Boolean constraint on whether the desired organizational unit is the [root](https://docs.aws.amazon.com/organizations/latest/userguide/orgs_getting-started_concepts.html#root) for the organization. For now, this is always `true`.

## Attributes Reference

The following attributes are exported:

* `arn` - The ARN of the organizational unit
* `id` - The ID of the organizational unit (`r-...`)
