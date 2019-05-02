---
layout: "aws"
page_title: "AWS: aws_appsync_resolver"
sidebar_current: "docs-aws-resource-appsync-resolver"
description: |-
  Provides an AppSync Resolver.
---

# Resource: aws_appsync_resolver

Provides an AppSync Resolver.

## Example Usage

```hcl
resource "aws_appsync_graphql_api" "test" {
  authentication_type = "API_KEY"
  name                = "tf-example"
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

resource "aws_appsync_datasource" "test" {
  api_id      = "${aws_appsync_graphql_api.test.id}"
  name        = "tf-example"
  type        = "HTTP"

  http_config {
    endpoint = "http://example.com"
  }
}

resource "aws_appsync_resolver" "test" {
  api_id           = "${aws_appsync_graphql_api.test.id}"
  field            = "singlePost"
  type             = "Query"
  data_source      = "${aws_appsync_datasource.test.name}"
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
}
```

## Argument Reference

The following arguments are supported:

* `api_id` - (Required) The API ID for the GraphQL API.
* `type` - (Required) The type name from the schema defined in the GraphQL API.
* `field` - (Required) The field name from the schema defined in the GraphQL API.
* `data_source` - (Required) The DataSource name.
* `request_template` - (Required) The request mapping template for this resolver.
* `response_template` - (Required) The response mapping template for this resolver.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - The ARN

## Import

`aws_appsync_resolver` can be imported with their `api_id`, a hyphen, `type`, a hypen and `field` e.g.

```
$ terraform import aws_appsync_resolver.example abcdef123456-exampleType-exampleField
```
