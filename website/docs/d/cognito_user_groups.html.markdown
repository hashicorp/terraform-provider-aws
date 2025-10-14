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

This data source supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
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
