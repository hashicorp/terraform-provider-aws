---
subcategory: "Cognito IDP (Identity Provider)"
layout: "aws"
page_title: "AWS: aws_cognito_user_pool_signing_certificate"
description: |-
  Get signing certificate of user pool
---

# Data Source: aws_cognito_user_pool_signing_certificate

Use this data source to get the signing certificate for a Cognito IdP user pool.

## Example Usage

```terraform
data "aws_cognito_user_pool_signing_certificate" "sc" {
  user_pool_id = aws_cognito_user_pool.my_pool.id
}
```

## Argument Reference

* `user_pool_id` - (Required) Cognito user pool ID.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `certificate` - Certificate string
