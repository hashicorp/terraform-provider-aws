layout: "aws"
page_title: "AWS: aws_cognito_user_pool_client"
side_bar_current: "docs-aws-resource-cognito-user-pool-client"
description: |-
  Provides a Cognito User Pool Client resource

# aws\_cognito\_user\_pool\_client

Provides a Cognito User Pool Client resource.

## Example Usage

### Create a basic user pool client

```hcl
resource "aws_cognito_user_pool" "pool" {
	name = "pool"
}

resource "aws_cognito_user_pool_client" "client" {
	name = "client"
	user_pool = "${aws_cognito_user_pool.pool.id}"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the user pool client.
* `user_pool` - (Required) The id of the user pool where you want to create a client.
* `generate_secret` - (Optional, Default: true) A boolean determining wether a secret key is generated.

## Attribute Reference

The following attributes are exported:

* `id` - The id of the user pool client.
* `secret` - The secret key for the user pool client.
