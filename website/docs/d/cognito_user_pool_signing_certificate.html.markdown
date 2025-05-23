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

This data source supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `user_pool_id` - (Required) Cognito user pool ID.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `certificate` - Certificate string
