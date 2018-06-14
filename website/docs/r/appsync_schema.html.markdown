---
layout: "aws"
page_title: "AWS: aws_appsync_schema"
sidebar_current: "docs-aws-resource-appsync-schema"
description: |-
  Provides an AppSync Schema.
---

# aws_appsync_schema

Provides an AppSync Schema.

## Example Usage

```hcl
resource "aws_appsync_graphql_api" "example" {
  authentication_type = "API_KEY"
  name = "tf_appsync_example"
}

resource "aws_appsync_schema" "example" {
  api_id = "${aws_appsync_graphql_api.example.id}"
  definition = <<EOF
schema {
	query: Query
}
type Query {
  test: Int
}
EOF
}
```

## Argument Reference

The following arguments are supported:

* `api_id` - (Required) The API ID for the GraphQL API for the DataSource.
* `definition` - (Required) The schema definition, in GraphQL schema language format
