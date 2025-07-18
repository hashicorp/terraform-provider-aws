---
subcategory: "AppSync"
layout: "aws"
page_title: "AWS: aws_appsync_event_api"
description: |-
  Manages an AWS AppSync Event API.
---

# Resource: aws_appsync_event_api

Manages an AWS AppSync Event API. Event APIs enable real-time subscriptions and event-driven communication in AppSync applications.

## Example Usage

### Basic Usage

```terraform
resource "aws_appsync_event_api" "example" {
  name = "example-event-api"
}
```

### With Owner Contact

```terraform
resource "aws_appsync_event_api" "example" {
  name          = "example-event-api"
  owner_contact = "admin@example.com"
}
```

### With Cognito Authentication

```terraform
resource "aws_cognito_user_pool" "example" {
  name = "example-user-pool"
}

resource "aws_appsync_event_api" "example" {
  name = "example-event-api"

  event_config {
    auth_providers {
      auth_type = "AMAZON_COGNITO_USER_POOLS"
      cognito_config {
        user_pool_id = aws_cognito_user_pool.example.id
        aws_region   = data.aws_region.current.name
      }
    }
  }
}

data "aws_region" "current" {}
```

### With Lambda Authorizer

```terraform
resource "aws_iam_role" "lambda" {
  name = "example-lambda-role"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRole"
        Effect = "Allow"
        Principal = {
          Service = "lambda.amazonaws.com"
        }
      }
    ]
  })
}

resource "aws_lambda_function" "example" {
  filename      = "lambda_authorizer.zip"
  function_name = "example-authorizer"
  role          = aws_iam_role.lambda.arn
  handler       = "index.handler"
  runtime       = "nodejs18.x"
}

resource "aws_appsync_event_api" "example" {
  name = "example-event-api"

  event_config {
    auth_providers {
      auth_type = "AWS_LAMBDA"
      lambda_authorizer_config {
        authorizer_uri                   = aws_lambda_function.example.invoke_arn
        authorizer_result_ttl_in_seconds = 300
      }
    }
  }
}
```

### With OpenID Connect

```terraform
resource "aws_appsync_event_api" "example" {
  name = "example-event-api"

  event_config {
    auth_providers {
      auth_type = "OPENID_CONNECT"
      openid_connect_config {
        issuer    = "https://example.com"
        client_id = "example-client-id"
      }
    }
  }
}
```

### With Authentication Modes

```terraform
resource "aws_appsync_event_api" "example" {
  name = "example-event-api"

  event_config {
    connection_auth_modes {
      auth_type = "API_KEY"
    }
    default_publish_auth_modes {
      auth_type = "API_KEY"
    }
    default_subscribe_auth_modes {
      auth_type = "API_KEY"
    }
  }
}
```

### With CloudWatch Logging

```terraform
resource "aws_iam_role" "cloudwatch" {
  name = "example-cloudwatch-role"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRole"
        Effect = "Allow"
        Principal = {
          Service = "appsync.amazonaws.com"
        }
      }
    ]
  })
}

resource "aws_iam_role_policy_attachment" "cloudwatch" {
  policy_arn = "arn:aws:iam::aws:policy/service-role/AppSyncPushToCloudWatchLogs"
  role       = aws_iam_role.cloudwatch.name
}

resource "aws_appsync_event_api" "example" {
  name = "example-event-api"

  event_config {
    log_config {
      cloudwatch_logs_role_arn = aws_iam_role.cloudwatch.arn
      log_level                = "ERROR"
    }
  }
}
```

## Argument Reference

The following arguments are required:

* `name` - (Required) Name of the Event API.

The following arguments are optional:

* `owner_contact` - (Optional) Contact information for the owner of the Event API.
* `event_config` - (Optional) Configuration for the Event API. See [Event Config](#event-config) below.
* `tags` - (Optional) Map of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

### Event Config

The `event_config` block supports the following:

* `auth_providers` - (Optional) List of authentication providers. See [Auth Providers](#auth-providers) below.
* `connection_auth_modes` - (Optional) List of authentication modes for connections. See [Auth Modes](#auth-modes) below.
* `default_publish_auth_modes` - (Optional) List of default authentication modes for publishing. See [Auth Modes](#auth-modes) below.
* `default_subscribe_auth_modes` - (Optional) List of default authentication modes for subscribing. See [Auth Modes](#auth-modes) below.
* `log_config` - (Optional) Logging configuration. See [Log Config](#log-config) below.

### Auth Providers

The `auth_providers` block supports the following:

* `auth_type` - (Required) Type of authentication provider. Valid values: `AMAZON_COGNITO_USER_POOLS`, `AWS_LAMBDA`, `OPENID_CONNECT`, `API_KEY`.
* `cognito_config` - (Optional) Configuration for Cognito user pool authentication. Required when `auth_type` is `AMAZON_COGNITO_USER_POOLS`. See [Cognito Config](#cognito-config) below.
* `lambda_authorizer_config` - (Optional) Configuration for Lambda authorization. Required when `auth_type` is `AWS_LAMBDA`. See [Lambda Authorizer Config](#lambda-authorizer-config) below.
* `openid_connect_config` - (Optional) Configuration for OpenID Connect. Required when `auth_type` is `OPENID_CONNECT`. See [OpenID Connect Config](#openid-connect-config) below.

### Cognito Config

The `cognito_config` block supports the following:

* `user_pool_id` - (Required) ID of the Cognito user pool.
* `aws_region` - (Required) AWS region where the user pool is located.
* `app_id_client_regex` - (Optional) Regular expression for matching the client ID.

### Lambda Authorizer Config

The `lambda_authorizer_config` block supports the following:

* `authorizer_uri` - (Required) URI of the Lambda function for authorization.
* `authorizer_result_ttl_in_seconds` - (Optional) TTL in seconds for the authorization result cache.
* `identity_validation_expression` - (Optional) Regular expression for identity validation.

### OpenID Connect Config

The `openid_connect_config` block supports the following:

* `issuer` - (Required) Issuer URL for the OpenID Connect provider.
* `client_id` - (Optional) Client ID for the OpenID Connect provider.
* `auth_ttl` - (Optional) TTL in seconds for the authentication token.
* `iat_ttl` - (Optional) TTL in seconds for the issued at time.

### Auth Modes

The `connection_auth_modes`, `default_publish_auth_modes`, and `default_subscribe_auth_modes` blocks support the following:

* `auth_type` - (Required) Type of authentication. Valid values: `API_KEY`, `AWS_IAM`, `AMAZON_COGNITO_USER_POOLS`, `OPENID_CONNECT`, `AWS_LAMBDA`.

### Log Config

The `log_config` block supports the following:

* `cloudwatch_logs_role_arn` - (Required) ARN of the IAM role for CloudWatch logs.
* `log_level` - (Required) Log level. Valid values: `NONE`, `ERROR`, `ALL`, `INFO`, `DEBUG`.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `api_id` - ID of the Event API.
* `arn` - ARN of the Event API.
* `created` - Date and time when the Event API was created.
* `dns` - DNS configuration for the Event API.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).
* `waf_web_acl_arn` - ARN of the associated WAF web ACL.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `30m`)
* `update` - (Default `30m`)
* `delete` - (Default `30m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import AppSync Event API using the `api_id`. For example:

```terraform
import {
  to = aws_appsync_event_api.example
  id = "example-api-id"
}
```

Using `terraform import`, import AppSync Event API using the `api_id`. For example:

```console
% terraform import aws_appsync_event_api.example example-api-id
```
