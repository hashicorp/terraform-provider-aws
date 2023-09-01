---
subcategory: "AppSync"
layout: "aws"
page_title: "AWS: aws_appsync_graphql_api"
description: |-
  Provides an AppSync GraphQL API.
---

# Resource: aws_appsync_graphql_api

Provides an AppSync GraphQL API.

## Example Usage

### API Key Authentication

```terraform
resource "aws_appsync_graphql_api" "example" {
  authentication_type = "API_KEY"
  name                = "example"
}
```

### AWS IAM Authentication

```terraform
resource "aws_appsync_graphql_api" "example" {
  authentication_type = "AWS_IAM"
  name                = "example"
}
```

### AWS Cognito User Pool Authentication

```terraform
resource "aws_appsync_graphql_api" "example" {
  authentication_type = "AMAZON_COGNITO_USER_POOLS"
  name                = "example"

  user_pool_config {
    aws_region     = data.aws_region.current.name
    default_action = "DENY"
    user_pool_id   = aws_cognito_user_pool.example.id
  }
}
```

### OpenID Connect Authentication

```terraform
resource "aws_appsync_graphql_api" "example" {
  authentication_type = "OPENID_CONNECT"
  name                = "example"

  openid_connect_config {
    issuer = "https://example.com"
  }
}
```

### AWS Lambda Authorizer Authentication

```terraform
resource "aws_appsync_graphql_api" "example" {
  authentication_type = "AWS_LAMBDA"
  name                = "example"

  lambda_authorizer_config {
    authorizer_uri = "arn:aws:lambda:us-east-1:123456789012:function:custom_lambda_authorizer"
  }
}

resource "aws_lambda_permission" "appsync_lambda_authorizer" {
  statement_id  = "appsync_lambda_authorizer"
  action        = "lambda:InvokeFunction"
  function_name = "custom_lambda_authorizer"
  principal     = "appsync.amazonaws.com"
  source_arn    = aws_appsync_graphql_api.example.arn
}
```

### With Multiple Authentication Providers

```terraform
resource "aws_appsync_graphql_api" "example" {
  authentication_type = "API_KEY"
  name                = "example"

  additional_authentication_provider {
    authentication_type = "AWS_IAM"
  }
}
```

### With Schema

```terraform
resource "aws_appsync_graphql_api" "example" {
  authentication_type = "AWS_IAM"
  name                = "example"

  schema = <<EOF
schema {
	query: Query
}
type Query {
  test: Int
}
EOF
}
```

### Enabling Logging

```terraform
data "aws_iam_policy_document" "assume_role" {
  statement {
    effect = "Allow"

    principals {
      type        = "Service"
      identifiers = ["appsync.amazonaws.com"]
    }

    actions = ["sts:AssumeRole"]
  }
}

resource "aws_iam_role" "example" {
  name               = "example"
  assume_role_policy = data.aws_iam_policy_document.assume_role.json
}

resource "aws_iam_role_policy_attachment" "example" {
  policy_arn = "arn:aws:iam::aws:policy/service-role/AWSAppSyncPushToCloudWatchLogs"
  role       = aws_iam_role.example.name
}

resource "aws_appsync_graphql_api" "example" {
  # ... other configuration ...

  log_config {
    cloudwatch_logs_role_arn = aws_iam_role.example.arn
    field_log_level          = "ERROR"
  }
}
```

### Associate Web ACL (v2)

```terraform
resource "aws_appsync_graphql_api" "example" {
  authentication_type = "API_KEY"
  name                = "example"
}

resource "aws_wafv2_web_acl_association" "example" {
  resource_arn = aws_appsync_graphql_api.example.arn
  web_acl_arn  = aws_wafv2_web_acl.example.arn
}

resource "aws_wafv2_web_acl" "example" {
  name        = "managed-rule-example"
  description = "Example of a managed rule."
  scope       = "REGIONAL"

  default_action {
    allow {}
  }

  rule {
    name     = "rule-1"
    priority = 1

    override_action {
      block {}
    }

    statement {
      managed_rule_group_statement {
        name        = "AWSManagedRulesCommonRuleSet"
        vendor_name = "AWS"
      }
    }

    visibility_config {
      cloudwatch_metrics_enabled = false
      metric_name                = "friendly-rule-metric-name"
      sampled_requests_enabled   = false
    }
  }

  visibility_config {
    cloudwatch_metrics_enabled = false
    metric_name                = "friendly-metric-name"
    sampled_requests_enabled   = false
  }
}
```

## Argument Reference

This resource supports the following arguments:

* `authentication_type` - (Required) Authentication type. Valid values: `API_KEY`, `AWS_IAM`, `AMAZON_COGNITO_USER_POOLS`, `OPENID_CONNECT`, `AWS_LAMBDA`
* `name` - (Required) User-supplied name for the GraphqlApi.
* `log_config` - (Optional) Nested argument containing logging configuration. Defined below.
* `openid_connect_config` - (Optional) Nested argument containing OpenID Connect configuration. Defined below.
* `user_pool_config` - (Optional) Amazon Cognito User Pool configuration. Defined below.
* `lambda_authorizer_config` - (Optional) Nested argument containing Lambda authorizer configuration. Defined below.
* `schema` - (Optional) Schema definition, in GraphQL schema language format. Terraform cannot perform drift detection of this configuration.
* `additional_authentication_provider` - (Optional) One or more additional authentication providers for the GraphqlApi. Defined below.
* `tags` - (Optional) Map of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.
* `xray_enabled` - (Optional) Whether tracing with X-ray is enabled. Defaults to false.
* `visibility` - (Optional) Sets the value of the GraphQL API to public (`GLOBAL`) or private (`PRIVATE`). If no value is provided, the visibility will be set to `GLOBAL` by default. This value cannot be changed once the API has been created.

### log_config

This argument supports the following arguments:

* `cloudwatch_logs_role_arn` - (Required) Amazon Resource Name of the service role that AWS AppSync will assume to publish to Amazon CloudWatch logs in your account.
* `field_log_level` - (Required) Field logging level. Valid values: `ALL`, `ERROR`, `NONE`.
* `exclude_verbose_content` - (Optional) Set to TRUE to exclude sections that contain information such as headers, context, and evaluated mapping templates, regardless of logging  level. Valid values: `true`, `false`. Default value: `false`

### additional_authentication_provider

This argument supports the following arguments:

* `authentication_type` - (Required) Authentication type. Valid values: `API_KEY`, `AWS_IAM`, `AMAZON_COGNITO_USER_POOLS`, `OPENID_CONNECT`, `AWS_LAMBDA`
* `openid_connect_config` - (Optional) Nested argument containing OpenID Connect configuration. Defined below.
* `user_pool_config` - (Optional) Amazon Cognito User Pool configuration. Defined below.

### openid_connect_config

This argument supports the following arguments:

* `issuer` - (Required) Issuer for the OpenID Connect configuration. The issuer returned by discovery MUST exactly match the value of iss in the ID Token.
* `auth_ttl` - (Optional) Number of milliseconds a token is valid after being authenticated.
* `client_id` - (Optional) Client identifier of the Relying party at the OpenID identity provider. This identifier is typically obtained when the Relying party is registered with the OpenID identity provider. You can specify a regular expression so the AWS AppSync can validate against multiple client identifiers at a time.
* `iat_ttl` - (Optional) Number of milliseconds a token is valid after being issued to a user.

### user_pool_config

This argument supports the following arguments:

* `default_action` - (Required only if Cognito is used as the default auth provider) Action that you want your GraphQL API to take when a request that uses Amazon Cognito User Pool authentication doesn't match the Amazon Cognito User Pool configuration. Valid: `ALLOW` and `DENY`
* `user_pool_id` - (Required) User pool ID.
* `app_id_client_regex` - (Optional) Regular expression for validating the incoming Amazon Cognito User Pool app client ID.
* `aws_region` - (Optional) AWS region in which the user pool was created.

### lambda_authorizer_config

This argument supports the following arguments:

* `authorizer_uri` - (Required) ARN of the Lambda function to be called for authorization. Note: This Lambda function must have a resource-based policy assigned to it, to allow `lambda:InvokeFunction` from service principal `appsync.amazonaws.com`.
* `authorizer_result_ttl_in_seconds` - (Optional) Number of seconds a response should be cached for. The default is 5 minutes (300 seconds). The Lambda function can override this by returning a `ttlOverride` key in its response. A value of 0 disables caching of responses. Minimum value of 0. Maximum value of 3600.
* `identity_validation_expression` - (Optional) Regular expression for validation of tokens before the Lambda function is called.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - API ID
* `arn` - ARN
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).
* `uris` - Map of URIs associated with the APIE.g., `uris["GRAPHQL"] = https://ID.appsync-api.REGION.amazonaws.com/graphql`

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import AppSync GraphQL API using the GraphQL API ID. For example:

```terraform
import {
  to = aws_appsync_graphql_api.example
  id = "0123456789"
}
```

Using `terraform import`, import AppSync GraphQL API using the GraphQL API ID. For example:

```console
% terraform import aws_appsync_graphql_api.example 0123456789
```
