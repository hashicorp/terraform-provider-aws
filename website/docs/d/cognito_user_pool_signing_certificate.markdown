---
subcategory: "Cognito"
layout: "aws"
page_title: "AWS: aws_cognito_user_pool_signing_certificate"
description: |-
  Get signing certificate of user pool
---

# Data Source: aws_cognito_user_pool_signing_certificate

Use this data source to get the signing certificate to the external SAML IdP.

## Example Usage

```terraform
data "aws_cognito_user_pool_signing_certificate" "sc" {
  user_pool_id = aws_cognito_user_pool.my_pool.id
}
```

## Argument Reference

* `user_pool_id` - (required) The Cognito user pool ids.


## Attributes Reference

* `certificate` - the certificate string
