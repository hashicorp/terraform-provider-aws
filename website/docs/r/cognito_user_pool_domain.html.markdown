---
layout: "aws"
page_title: "AWS: aws_cognito_user_pool_domain"
side_bar_current: "docs-aws-resource-cognito-user-pool-domain"
description: |-
  Provides a Cognito User Pool Domain resource.
---

# aws_cognito_user_pool_domain

Provides a Cognito User Pool Domain resource.

## Example Usage

```hcl
resource "aws_cognito_user_pool" "pool" {
  name = "mypool"
}

resource "aws_cognito_user_pool_domain" "domain" {
  domain = "mydomain"
  user_pool_id = "${aws_cognito_user_pool.pool.id}"
}
```

## Argument Reference

The following arguments are supported:

* `domain` - (Required) The domain string.
* `user_pool_id` - (Required) The user pool ID.
