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

## Argument Reference

The following arguments are supported:

* `api_id` - (Required) The ID of the associated AppSync API.
* `data_source` - (Required) The Function data source name.
* `max_batch_size` - (Optional) The maximum batching size for a resolver. Valid values are between `0` and `2000`.
* `name` - (Required) The Function name. The function name does not have to be unique.
* `request_mapping_template` - (Required) The Function request mapping template. Functions support only the 2018-05-29 version of the request mapping template.
* `response_mapping_template` - (Required) The Function response mapping template.
* `description` - (Optional) The Function description.
* `sync_config` - (Optional) Describes a Sync configuration for a resolver. See [Sync Config](#sync-config).
* `function_version` - (Optional) The version of the request mapping template. Currently the supported value is `2018-05-29`.

### Sync Config

The following arguments are supported:

* `conflict_detection` - (Optional) The Conflict Detection strategy to use. Valid values are `NONE` and `VERSION`.
* `conflict_handler` - (Optional) The Conflict Resolution strategy to perform in the event of a conflict. Valid values are `NONE`, `OPTIMISTIC_CONCURRENCY`, `AUTOMERGE`, and `LAMBDA`.
* `lambda_conflict_handler_config` - (Optional) The Lambda Conflict Handler Config when configuring `LAMBDA` as the Conflict Handler. See [Lambda Conflict Handler Config](#lambda-conflict-handler-config).

#### Lambda Conflict Handler Config

The following arguments are supported:

* `lambda_conflict_handler_arn` - (Optional) The Amazon Resource Name (ARN) for the Lambda function to use as the Conflict Handler.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - API Function ID (Formatted as ApiId-FunctionId)
* `arn` - The ARN of the Function object.
* `function_id` - A unique ID representing the Function object.

## Import

`aws_appsync_function` can be imported using the AppSync API ID and Function ID separated by `-`, e.g.,

```
$ terraform import aws_appsync_function.example xxxxx-yyyyy
```
