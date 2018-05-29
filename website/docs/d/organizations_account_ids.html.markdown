---
layout: "aws"
page_title: "AWS: aws_organizations_account_ids"
sidebar_current: "docs-aws-datasource-organizations_account_ids"
description: |-
  Provides a list of AWS Account IDs in the joined AWS Organization.
---

# Data Source: aws_organizations_account_ids

Use this data source to get a list of AWS Account IDs in the joined organization.

~> **NOTE:** Account management must be done from the organization's master account.

## Example Usage

```hcl
data "aws_organizations_account_ids" "accounts" {}

resource "aws_ami_copy" "example" {
  name              = "terraform-example"
  description       = "A copy of ami-f2d3638a"
  source_ami_id     = "ami-f2d3638a"
  source_ami_region = "us-west-2"
}

resource "aws_ami_launch_permission" "example1" {
  count      = "${length(data.aws_organizations_account_ids.accounts.ids)}"
  image_id   = "${aws_ami_copy.example.id}"
  account_id = "${data.aws_organizations_account_ids.accounts.ids[count.index]}"
}
```

## Argument Reference

There are no arguments available for this data source.

## Attributes Reference

`ids` is set to the list of Account IDs
