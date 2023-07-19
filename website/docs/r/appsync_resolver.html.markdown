---
subcategory: "AppSync"
layout: "aws"
page_title: "AWS: aws_appsync_resolver"
description: |-
  Provides an AppSync Resolver.
---

# Resource: aws_appsync_resolver

Provides an AppSync Resolver.

## Example Usage

```terraform
resource "aws_appsync_graphql_api" "test" {
  authentication_type = "API_KEY"
  name                = "tf-example"

  schema = <<EOF
type Mutation {
	putPost(id: ID!, title: String!): Post
}

type Post {
	id: ID!
	title: String!
}

type Query {
	singlePost(id: ID!): Post
}

schema {
	query: Query
	mutation: Mutation
}
EOF
}

resource "aws_appsync_datasource" "test" {
  api_id = aws_appsync_graphql_api.test.id
  name   = "tf_example"
  type   = "HTTP"

  http_config {
    endpoint = "http://example.com"
  }
}

# UNIT type resolver (default)
resource "aws_appsync_resolver" "test" {
  api_id      = aws_appsync_graphql_api.test.id
  field       = "singlePost"
  type        = "Query"
  data_source = aws_appsync_datasource.test.name

  request_template = <<EOF
{
    "version": "2018-05-29",
    "method": "GET",
    "resourcePath": "/",
    "params":{
        "headers": $utils.http.copyheaders($ctx.request.headers)
    }
}
EOF

  response_template = <<EOF
#if($ctx.result.statusCode == 200)
    $ctx.result.body
#else
    $utils.appendError($ctx.result.body, $ctx.result.statusCode)
#end
EOF

  caching_config {
    caching_keys = [
      "$context.identity.sub",
      "$context.arguments.id",
    ]
    ttl = 60
  }
}

# PIPELINE type resolver
resource "aws_appsync_resolver" "Mutation_pipelineTest" {
  type              = "Mutation"
  api_id            = aws_appsync_graphql_api.test.id
  field             = "pipelineTest"
  request_template  = "{}"
  response_template = "$util.toJson($ctx.result)"
  kind              = "PIPELINE"
  pipeline_config {
    functions = [
      aws_appsync_function.test1.function_id,
      aws_appsync_function.test2.function_id,
      aws_appsync_function.test3.function_id,
    ]
  }
}
```

## Example Usage JS

```terraform
resource "aws_appsync_resolver" "example" {
  type   = "Query"
  api_id = aws_appsync_graphql_api.test.id
  field  = "pipelineTest"
  kind   = "PIPELINE"
  code   = file("some-code-dir")

  runtime {
    name            = "APPSYNC_JS"
    runtime_version = "1.0.0"
  }

  pipeline_config {
    functions = [
      aws_appsync_function.test.function_id,
    ]
  }
}
```

## Argument Reference

This resource supports the following arguments:

* `api_id` - (Required) API ID for the GraphQL API.
* `code` - (Optional) The function code that contains the request and response functions. When code is used, the runtime is required. The runtime value must be APPSYNC_JS.
* `type` - (Required) Type name from the schema defined in the GraphQL API.
* `field` - (Required) Field name from the schema defined in the GraphQL API.
* `request_template` - (Optional) Request mapping template for UNIT resolver or 'before mapping template' for PIPELINE resolver. Required for non-Lambda resolvers.
* `response_template` - (Optional) Response mapping template for UNIT resolver or 'after mapping template' for PIPELINE resolver. Required for non-Lambda resolvers.
* `data_source` - (Optional) Data source name.
* `max_batch_size` - (Optional) Maximum batching size for a resolver. Valid values are between `0` and `2000`.
* `kind`  - (Optional) Resolver type. Valid values are `UNIT` and `PIPELINE`.
* `sync_config` - (Optional) Describes a Sync configuration for a resolver. See [Sync Config](#sync-config).
* `pipeline_config` - (Optional) The caching configuration for the resolver. See [Pipeline Config](#pipeline-config).
* `caching_config` - (Optional) The Caching Config. See [Caching Config](#caching-config).
* `runtime` - (Optional) Describes a runtime used by an AWS AppSync pipeline resolver or AWS AppSync function. Specifies the name and version of the runtime to use. Note that if a runtime is specified, code must also be specified. See [Runtime](#runtime).

### Caching Config

* `caching_keys` - (Optional) The caching keys for a resolver that has caching activated. Valid values are entries from the $context.arguments, $context.source, and $context.identity maps.
* `ttl` - (Optional) The TTL in seconds for a resolver that has caching activated. Valid values are between `1` and `3600` seconds.

### Pipeline Config

* `functions` - (Optional) A list of Function objects.

### Sync Config

* `conflict_detection` - (Optional) Conflict Detection strategy to use. Valid values are `NONE` and `VERSION`.
* `conflict_handler` - (Optional) Conflict Resolution strategy to perform in the event of a conflict. Valid values are `NONE`, `OPTIMISTIC_CONCURRENCY`, `AUTOMERGE`, and `LAMBDA`.
* `lambda_conflict_handler_config` - (Optional) Lambda Conflict Handler Config when configuring `LAMBDA` as the Conflict Handler. See [Lambda Conflict Handler Config](#lambda-conflict-handler-config).

#### Lambda Conflict Handler Config

* `lambda_conflict_handler_arn` - (Optional) ARN for the Lambda function to use as the Conflict Handler.

### Runtime

* `name` - (Optional) The name of the runtime to use. Currently, the only allowed value is `APPSYNC_JS`.
* `runtime_version` - (Optional) The version of the runtime to use. Currently, the only allowed version is `1.0.0`.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import `aws_appsync_resolver` using the `api_id`, a hyphen, `type`, a hypen and `field`. For example:

```terraform
import {
  to = aws_appsync_resolver.example
  id = "abcdef123456-exampleType-exampleField"
}
```

Using `terraform import`, import `aws_appsync_resolver` using the `api_id`, a hyphen, `type`, a hypen and `field`. For example:

```console
% terraform import aws_appsync_resolver.example abcdef123456-exampleType-exampleField
```
