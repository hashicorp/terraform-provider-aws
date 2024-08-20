---
subcategory: "AppSync"
layout: "aws"
page_title: "AWS: aws_appsync_function"
description: |-
  Provides an AppSync Function.
---

# Resource: aws_appsync_function

Provides an AppSync Function.

## Example Usage

```terraform
resource "aws_appsync_graphql_api" "example" {
  authentication_type = "API_KEY"
  name                = "example"
  schema              = <<EOF
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

resource "aws_appsync_datasource" "example" {
  api_id = aws_appsync_graphql_api.example.id
  name   = "example"
  type   = "HTTP"

  http_config {
    endpoint = "http://example.com"
  }
}

resource "aws_appsync_function" "example" {
  api_id                   = aws_appsync_graphql_api.example.id
  data_source              = aws_appsync_datasource.example.name
  name                     = "example"
  request_mapping_template = <<EOF
{
    "version": "2018-05-29",
    "method": "GET",
    "resourcePath": "/",
    "params":{
        "headers": $utils.http.copyheaders($ctx.request.headers)
    }
}
EOF

  response_mapping_template = <<EOF
#if($ctx.result.statusCode == 200)
    $ctx.result.body
#else
    $utils.appendError($ctx.result.body, $ctx.result.statusCode)
#end
EOF
}
```

## Example Usage With Code

```terraform
resource "aws_appsync_function" "example" {
  api_id      = aws_appsync_graphql_api.example.id
  data_source = aws_appsync_datasource.example.name
  name        = "example"
  code        = file("some-code-dir")

  runtime {
    name            = "APPSYNC_JS"
    runtime_version = "1.0.0"
  }
}
```

## Argument Reference

This resource supports the following arguments:

* `api_id` - (Required) ID of the associated AppSync API.
* `code` - (Optional) The function code that contains the request and response functions. When code is used, the runtime is required. The runtime value must be APPSYNC_JS.
* `data_source` - (Required) Function data source name.
* `max_batch_size` - (Optional) Maximum batching size for a resolver. Valid values are between `0` and `2000`.
* `name` - (Required) Function name. The function name does not have to be unique.
* `request_mapping_template` - (Optional) Function request mapping template. Functions support only the 2018-05-29 version of the request mapping template.
* `response_mapping_template` - (Optional) Function response mapping template.
* `description` - (Optional) Function description.
* `runtime` - (Optional) Describes a runtime used by an AWS AppSync pipeline resolver or AWS AppSync function. Specifies the name and version of the runtime to use. Note that if a runtime is specified, code must also be specified. See [`runtime` Block](#runtime-block) for details.
* `sync_config` - (Optional) Describes a Sync configuration for a resolver. See [`sync_config` Block](#sync_config-block) for details.
* `function_version` - (Optional) Version of the request mapping template. Currently the supported value is `2018-05-29`. Does not apply when specifying `code`.

### `runtime` Block

The `runtime` configuration block supports the following arguments:

* `name` - (Optional) The name of the runtime to use. Currently, the only allowed value is `APPSYNC_JS`.
* `runtime_version` - (Optional) The version of the runtime to use. Currently, the only allowed version is `1.0.0`.

### `sync_config` Block

The `sync_config` configuration block supports the following arguments:

* `conflict_detection` - (Optional) Conflict Detection strategy to use. Valid values are `NONE` and `VERSION`.
* `conflict_handler` - (Optional) Conflict Resolution strategy to perform in the event of a conflict. Valid values are `NONE`, `OPTIMISTIC_CONCURRENCY`, `AUTOMERGE`, and `LAMBDA`.
* `lambda_conflict_handler_config` - (Optional) Lambda Conflict Handler Config when configuring `LAMBDA` as the Conflict Handler. See [`lambda_conflict_handler_config` Block](#lambda_conflict_handler_config-block) for details.

#### `lambda_conflict_handler_config` Block

The `lambda_conflict_handler_config` configuration block supports the following arguments:

* `lambda_conflict_handler_arn` - (Optional) ARN for the Lambda function to use as the Conflict Handler.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - API Function ID (Formatted as ApiId-FunctionId)
* `arn` - ARN of the Function object.
* `function_id` - Unique ID representing the Function object.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import `aws_appsync_function` using the AppSync API ID and Function ID separated by `-`. For example:

```terraform
import {
  to = aws_appsync_function.example
  id = "xxxxx-yyyyy"
}
```

Using `terraform import`, import `aws_appsync_function` using the AppSync API ID and Function ID separated by `-`. For example:

```console
% terraform import aws_appsync_function.example xxxxx-yyyyy
```
