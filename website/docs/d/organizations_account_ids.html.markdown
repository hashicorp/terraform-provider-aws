---
layout: "aws"
page_title: "AWS: aws_organizations_account_ids"
sidebar_current: "docs-aws-datasource-organizations-account-ids"
description: |-
    Provides a list of Account IDs in an Organization or Organizational Unit.Use this data source to get a list of Account IDs in an Organization or Organizational Unit
---

# Data Source: aws_organizations_account_ids

`aws_organizations_account_ids` provides a list of AccountIds in an [organization](https://docs.aws.amazon.com/organizations/latest/userguide/orgs_manage_org.html) or [organizational unit](https://docs.aws.amazon.com/organizations/latest/userguide/orgs_manage_ous.html).

Will give an error if Organizations aren't enabled - see `aws_organizations_organization`.

## Example Usage

```hcl
data "aws_organizations_account_ids" "master" {}
```

## Argument Reference

* `parent_id` - (Optional) The ID for the [parent root](https://docs.aws.amazon.com/organizations/latest/userguide/orgs_getting-started_concepts.html#root) or [organizational unit](https://docs.aws.amazon.com/organizations/latest/userguide/orgs_getting-started_concepts.html#organizationalunit) whose accounts you want to list.  If you specify the root you get the list of all the accounts that are not in any organizational unit.  If you specify an organizational unit, you get a list of all the accounts in only that organizational unit, and not any child organizational units.

## Attributes Reference

The following attributes are exported:

* `ids` - is set to a list Account IDs.
