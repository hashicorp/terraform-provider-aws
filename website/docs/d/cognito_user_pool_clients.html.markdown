---
subcategory: "Cognito IDP (Identity Provider)"
layout: "aws"
page_title: "AWS: aws_cognito_user_pool_clients"
description: |-
  Get list of cognito user pool clients connected to user pool.
---

# Data Source: aws_cognito_user_pool_clients

Use this data source to get a list of Cognito user pools clients for a Cognito IdP user pool.

## Example Usage

```terraform
data "aws_cognito_user_pool_clients" "main" {
  user_pool_id = aws_cognito_user_pool.main.id
}
```

## Argument Reference

This data source supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `user_pool_id` - (Required) Cognito user pool ID.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `client_ids` - List of Cognito user pool client IDs.
* `client_names` - List of Cognito user pool client names.
