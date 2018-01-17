---
layout: "aws"
page_title: "AWS: aws_cognito_user_pool_group"
side_bar_current: "docs-aws-resource-cognito-user-pool-group"
description: |-
  Provides a Cognito User Pool Group resource.
---

# aws_cognito_user_pool_group

Provides a Cognito User Pool Group resource.

## Example Usage

### Create a basic user pool group

```hcl
resource "aws_cognito_user_pool" "pool" {
  name = "pool"
}

resource "aws_cognito_user_pool_group" "group" {
  name = "group"

  user_pool_id = "${aws_cognito_user_pool.pool.id}"
}
```

## Argument Reference

The following arguments are supported:

* `description` - (Optional) A string containing the description of the group.
* `name` - (Required) The name of the group.
* `precedence` - (Optional) The precedence of this group relative to the other groups in the pool.
* `role_arn` - (Optional) The role ARN for the group.
* `user_pool_id` - (Required) The user pool ID for the user pool.
