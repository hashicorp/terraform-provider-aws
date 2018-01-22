---
layout: "aws"
page_title: "AWS: aws_organization_attributes"
sidebar_current: "docs-aws-datasource-orgnization-attributes"
description: |-
  Get AWS Organization attributes
---

# aws_organization_attributes

Use this data source to lookup current attributes for your AWS Organization

## Example Usage

```hcl
data "aws_organization_attributes" "example" {}
```

## Attributes Reference

* `arn` - The Amazon Resource Name (ARN) of an organization
* `id` - The unique identifier (ID) of an organization
* `master_account_arn` - The Amazon Resource Name (ARN) of the account that is designated as the master account for the organization
* `master_account_email` - The email address that is associated with the AWS account that is designated as the master account for the organization
* `master_account_id` - The unique identifier (ID) of the master account of an organization
* `feature_set` - The functionality that currently is available to the organization
* `available_policy_types` - A list of policy types mappings that are enabled for this organization
  * `status` -  The status of the policy type as it relates to the associated root
  * `type` - The name of the policy type