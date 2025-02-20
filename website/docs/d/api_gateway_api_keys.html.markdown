---
subcategory: "API Gateway"
layout: "aws"
page_title: "AWS: aws_apigateway_api_keys"
description: |-
  Terraform data source for managing an AWS API Gateway API Keys.
---

# Data Source: aws_apigateway_api_keys

Use this data source to get pre-existing API Key list.

## Example Usage

```terraform
data "aws_api_gateway_api_keys" "my_api_keys" {}
```

## Argument Reference

This resource supports the following arguments:

* `customer_id` - (Optional) Amazon Web Services Marketplace customer identifier, when integrating with the Amazon Web Services SaaS Marketplace.

* `include_values` - (Optional) Set this value to `true` if you wish the result contains the key value. Defaults to `false`.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `id` - AWS Region.
* `items` - List of objects containing API Key information. See below.

### `items`

* `id` - Set to the ID of the API Key.
* `name` - Set to the name of the API Key.
* `value` - Set to the value of the API Key.
* `created_date` - Date and time when the API Key was created.
* `last_updated_date` - Date and time when the API Key was last updated.
* `customer_id` - Amazon Web Services Marketplace customer identifier, when integrating with the Amazon Web Services SaaS Marketplace.
* `description` - Description of the API Key.
* `enabled` - Whether the API Key is enabled.
* `tags` - Map of tags for the resource.
