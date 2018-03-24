---
layout: "aws"
page_title: "AWS: aws_appsync_api_key"
sidebar_current: "docs-aws-resource-appsync-api-key"
description: |-
  Provides an AppSync API Key.
---

# aws_appsync_api_key

Provides an AppSync API Key.

## Example Usage

```hcl
resource "aws_appsync_graphql_api" "example" {
  authentication_type = "API_KEY"
  name = "example"
}
resource "aws_appsync_api_key" "self" {
 appsync_api_id = "${aws_appsync_graphql_api.example.id}"
 validity_period_days = 364
}
resource "aws_appsync_api_key" "partner" {
 appsync_api_id = "${aws_appsync_graphql_api.example.id}"
 valid_till_date = "30/11/2018"
}
```

## Argument Reference

The following arguments are supported:

* `api_id` - (Required) The ID of the associated AppSync API
* `description` - (Optional) The API key description. Defaults to "Managed by Terraform".
* `expires` - (Optional)  RFC3339 string representation of the expiry date. By default, it is 7 days from the date of creation.


## Attributes Reference

The following attributes are exported:

* `id` - API ID
* `key` - The API key
