---
subcategory: "AppSync"
layout: "aws"
page_title: "AWS: aws_appsync_type"
description: |-
  Provides an AppSync Type.
---

# Resource: aws_appsync_type

Provides an AppSync Type.

## Example Usage

```terraform
resource "aws_appsync_graphql_api" "example" {
  authentication_type = "API_KEY"
  name                = "example"
}

resource "aws_appsync_type" "example" {
  api_id     = aws_appsync_graphql_api.example.id
  format     = "SDL"
  definition = <<EOF
type Mutation

{
putPost(id: ID!,title: String! ): Post

}
EOF  
}
```

## Argument Reference

The following arguments are supported:

* `api_id` - (Required) GraphQL API ID.
* `format` - (Required) The type format: `SDL` or `JSON`.
* `definition` - (Required) The type definition.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - The ARN of the type.
* `description` - The type description.
* `id` - The ID is constructed from `api-id:format:name`.
* `name` - The type name.

## Import

Appsync Types can be imported using the `id` e.g.,

```
$ terraform import aws_appsync_type.example api-id:format:name
```
