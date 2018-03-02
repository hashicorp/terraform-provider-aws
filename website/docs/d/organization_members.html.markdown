---
layout: "aws"
page_title: "AWS: aws_organization_members"
sidebar_current: "docs-aws-datasource-organization-members"
description: |-
  To retrieve a list of all of the accounts in an organization
---

# Data Source: aws_organization_members

Use this data source to lookup the list of all of the accounts in an organization.  This operation can be called only from the organization's master account.

## Example Usage

```hcl
data "aws_organization_members" "main" {}

resource "aws_ami_launch_permission" "base-ami" {
  count      = "${data.aws_organization_members.main.total_accounts}"
  image_id   = "${aws_ami_from_instance.ecs.id}"
  account_id = "${element(data.aws_organization_members.main.ids, count.index)}"
}
```

## Argument Reference

There are no arguments available for this data source.

## Attributes Reference

* `names` - A list of the names for each of the AWS accounts under the organization
* `ids` - A list of the account IDs for each of the AWS accounts under the organization
* `arns` - A list of the ARNs for each of the AWS accounts under the organization
* `emails` - A list of the emails for each of the AWS accounts under the organization
* `accounts` - A list of maps with the data returned by the CLi command: aws organizations list-accounts
* `total_accounts` - The total number of AWS accounts under the organization
