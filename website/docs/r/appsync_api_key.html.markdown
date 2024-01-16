---
subcategory: "AppSync"
layout: "aws"
page_title: "AWS: aws_appsync_api_key"
description: |-
  Provides an AppSync API Key.
---

# Resource: aws_appsync_api_key

Provides an AppSync API Key.

## Example Usage

```terraform
resource "aws_appsync_graphql_api" "example" {
  authentication_type = "API_KEY"
  name                = "example"
}

resource "aws_appsync_api_key" "example" {
  api_id  = aws_appsync_graphql_api.example.id
  expires = "2018-05-03T04:00:00Z"
}
```

## Argument Reference

This resource supports the following arguments:

* `api_id` - (Required) ID of the associated AppSync API
* `description` - (Optional) API key description. Defaults to "Managed by Terraform".
* `expires` - (Optional) RFC3339 string representation of the expiry date. Rounded down to nearest hour. By default, it is 7 days from the date of creation.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - API Key ID (Formatted as ApiId:Key)
* `key` - API key

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import `aws_appsync_api_key` using the AppSync API ID and key separated by `:`. For example:

```terraform
import {
  to = aws_appsync_api_key.example
  id = "xxxxx:yyyyy"
}
```

Using `terraform import`, import `aws_appsync_api_key` using the AppSync API ID and key separated by `:`. For example:

```console
% terraform import aws_appsync_api_key.example xxxxx:yyyyy
```
