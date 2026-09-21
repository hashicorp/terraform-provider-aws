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
* `definition` - (Required) Type definition.
* `format` - (Required) Type format: `SDL` or `JSON`.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the type.
* `description` - Type description.
* `id` - ID constructed from `api-id:format:name`.
* `name` - Type name.

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
