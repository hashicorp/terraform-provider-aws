---
subcategory: "Cognito IDP (Identity Provider)"
layout: "aws"
page_title: "AWS: aws_cognito_user_pool_session_configuration"
description: |-
  Manages session tracking for an AWS Cognito User Pool.
---

# Resource: aws_cognito_user_pool_session_configuration

Manages session tracking for an AWS Cognito User Pool, including the maximum number of active token sessions that each user can have. For more information, see [SetUserPoolSessionConfiguration](https://docs.aws.amazon.com/cognito-user-identity-pools/latest/APIReference/API_SetUserPoolSessionConfiguration.html) in the Amazon Cognito API Reference.

~> **NOTE:** Session tracking requires a user pool on the Essentials or Plus [feature plan](https://docs.aws.amazon.com/cognito/latest/developerguide/cognito-sign-in-feature-plans.html).

~> **NOTE:** Cognito has no API to remove a session configuration. Destroying this resource resets the user pool to the defaults: session tracking `DISABLED` and `max_per_user` of `20`.

## Example Usage

```terraform
resource "aws_cognito_user_pool" "example" {
  name           = "example"
  user_pool_tier = "ESSENTIALS"
}

resource "aws_cognito_user_pool_session_configuration" "example" {
  user_pool_id          = aws_cognito_user_pool.example.id
  user_session_tracking = "ENABLED"

  token_session {
    max_per_user = 10
  }
}
```

## Argument Reference

The following arguments are required:

* `token_session` - (Required) Pool-wide token session settings. See [`token_session`](#token_session) below.
* `user_pool_id` - (Required) ID of the user pool.
* `user_session_tracking` - (Required) Whether Cognito tracks sessions for users in the user pool. Valid values are `ENABLED` and `DISABLED`. While tracking is `DISABLED`, app client session configurations are kept but have no effect.

The following arguments are optional:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

### `token_session`

* `max_per_user` - (Required) Maximum number of active token sessions that a user can have across the user pool. Valid values are between `1` and `100`. When a user reaches this limit, their next successful sign-in evicts their least recently used token session. Lowering the value doesn't evict existing sessions until the user's next sign-in.

## Attribute Reference

This resource exports no additional attributes.

## Import

In Terraform v1.12.0 and later, the [`import` block](https://developer.hashicorp.com/terraform/language/import) can be used with the `identity` attribute. For example:

```terraform
import {
  to = aws_cognito_user_pool_session_configuration.example
  identity = {
    user_pool_id = "us-west-2_example123"
  }
}

resource "aws_cognito_user_pool_session_configuration" "example" {
  ### Configuration omitted for brevity ###
}
```

### Identity Schema

#### Required

* `user_pool_id` (String) ID of the Cognito User Pool.

#### Optional

* `account_id` (String) AWS Account where this resource is managed.
* `region` (String) Region where this resource is managed.

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import a Cognito User Pool Session Configuration using the `user_pool_id`. For example:

```terraform
import {
  to = aws_cognito_user_pool_session_configuration.example
  id = "us-west-2_example123"
}
```

Using `terraform import`, import a Cognito User Pool Session Configuration using the `user_pool_id`. For example:

```console
% terraform import aws_cognito_user_pool_session_configuration.example us-west-2_example123
```
