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

This resource supports the following arguments:

* `api_id` - (Required) GraphQL API ID.
* `format` - (Required) The type format: `SDL` or `JSON`.
* `definition` - (Required) The type definition.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - The ARN of the type.
* `description` - The type description.
* `id` - The ID is constructed from `api-id:format:name`.
* `name` - The type name.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Appsync Types using the `id`. For example:

```terraform
import {
  to = aws_appsync_type.example
  id = "api-id:format:name"
}
```

Using `terraform import`, import Appsync Types using the `id`. For example:

```console
% terraform import aws_appsync_type.example api-id:format:name
```
