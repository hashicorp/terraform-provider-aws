---
subcategory: "Cognito IDP (Identity Provider)"
layout: "aws"
page_title: "AWS: aws_cognito_user_pools"
description: |-
  Get list of cognito user pools.
---

# Data Source: aws_cognito_user_pools

Use this data source to get a list of cognito user pools.

## Example Usage

```terraform
data "aws_api_gateway_rest_api" "selected" {
  name = var.api_gateway_name
}

data "aws_cognito_user_pools" "selected" {
  name = var.cognito_user_pool_name
}

resource "aws_api_gateway_authorizer" "cognito" {
  name          = "cognito"
  type          = "COGNITO_USER_POOLS"
  rest_api_id   = data.aws_api_gateway_rest_api.selected.id
  provider_arns = data.aws_cognito_user_pools.selected.arns
}
```

## Argument Reference

* `name` - (Required) Name of the cognito user pools. Name is not a unique attribute for cognito user pool, so multiple pools might be returned with given name. If the pool name is expected to be unique, you can reference the pool id via ```tolist(data.aws_cognito_user_pools.selected.ids)[0]```

## Attributes Reference

* `ids` - Set of cognito user pool ids.
* `arns` - Set of cognito user pool Amazon Resource Names (ARNs).
