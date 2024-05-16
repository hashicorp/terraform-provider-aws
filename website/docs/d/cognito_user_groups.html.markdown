---
subcategory: "Cognito IDP (Identity Provider)"
layout: "aws"
page_title: "AWS: aws_cognito_user_groups"
description: |-
  Terraform data source for managing AWS Cognito IDP (Identity Provider) User Groups.
---

# Data Source: aws_cognito_user_groups

Terraform data source for managing AWS Cognito IDP (Identity Provider) User Groups.

## Example Usage

### Basic Usage

```terraform
data "aws_cognito_user_groups" "example" {
  user_pool_id = "us-west-2_aaaaaaaaa"
}
```

## Argument Reference

The following arguments are required:

* `user_pool_id` - (Required) User pool the client belongs to.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `id` - User pool identifier.
* `groups` - List of groups. See [`groups`](#groups) below.

### groups

* `description` - Description of the user group.
* `group_name` - Name of the user group.
* `precedence` - Precedence of the user group.
* `role_arn` - ARN of the IAM role to be associated with the user group.
