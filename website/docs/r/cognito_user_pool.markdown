layout: "aws"
page_title: "AWS: aws_cognito_user_pool"
side_bar_current: "docs-aws-resource-cognito-user-pool"
description: |-
  Provides a Cognito User Pool resource.

# aws_cognito_user_pool

Provides a Cognito User Pool resource.

## Example Usage

### Create a basic user pool

```hcl
resource "aws_cognito_user_pool" "pool" {
  name = "pool"
}
```

### Create a user pool and with custom messages
```hcl
resource "aws_cognito_user_pool" "pool" {
  name = "pool"

  email_verification_subject = "Device Verification Code"
  email_verification_message = <<MSG
If you requested to have this device verify
please use the following code {####}. Otherwise
ignore this message
Have a nice day!
MSG
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the user pool.
* `email_verification_subject` - (Optional) The subject line for verification emails.
* `email_verification_message` - (Optional) The message body for verification emails. Must contain {####} placeholder.
* `mfa_configuration` - (Optional, Default: OFF) Set to enable multi-factor authentication. Must be one of the following values (ON, OFF, OPTIONAL)

## Attribute Reference

The following attributes are exported:

* `id` - The id of the user pool.
