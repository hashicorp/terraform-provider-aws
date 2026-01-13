---
subcategory: "Cognito IDP (Identity Provider)"
layout: "aws"
page_title: "AWS: aws_cognito_user_group"
description: |-
  Terraform data source for managing an AWS Cognito IDP (Identity Provider) User Group.
---

# Data Source: aws_cognito_user_group

Terraform data source for managing an AWS Cognito IDP (Identity Provider) User Group.

## Example Usage

### Basic Usage

```terraform
data "aws_cognito_user_group" "example" {
  user_pool_id = "us-west-2_aaaaaaaaa"
  name         = "example"
}
```

## Argument Reference

This data source supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `name` - (Required) Name of the user group.
* `user_pool_id` - (Required) User pool the client belongs to.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `description` - Description of the user group.
* `id` - A comma-delimited string concatenating `name` and `user_pool_id`.
* `precedence` - Precedence of the user group.
* `role_arn` - ARN of the IAM role to be associated with the user group.
