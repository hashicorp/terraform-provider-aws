---
subcategory: "Cognito"
layout: "aws"
page_title: "AWS: aws_cognito_user_pool_clients"
description: |-
  Get list of cognito user pool clients connected to user pool.
---

# Data Source: aws_cognito_user_pool_clients

Use this data source to get a list of cognito user pools clients on a given userpool.

## Example Usage

```terraform
data "aws_cognito_user_pool_clients" "main" {
  user_pool_id = aws_cognito_user_pool.main.id
}
```

## Argument Reference

* `user_pool_id` - (required) The Cognito user pool id.


## Attributes Reference

* `client_ids` - The set of cognito user pool client ids.
