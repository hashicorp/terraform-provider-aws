---
subcategory: "API Gateway"
layout: "aws"
page_title: "AWS: aws_api_gateway_api_keys"
description: |-
  Terraform data source for managing AWS API Gateway API Keys.
---

# Data Source: aws_api_gateway_api_keys

Terraform data source for managing AWS API Gateway API Keys.

## Example Usage

```terraform
data "aws_api_gateway_api_keys" "example" {}
```

## Argument Reference

This data source supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `customer_id` - (Optional) Amazon Web Services Marketplace customer identifier, when integrating with the Amazon Web Services SaaS Marketplace.
* `include_values` - (Optional) Set this value to `true` if you wish the result contains the key value. Defaults to `false`.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `id` - AWS Region.
* `items` - List of objects containing API Key information. See below.

### `items`

* `id` - ID of the API Key.
* `name` - Name of the API Key.
* `value` - Value of the API Key.
* `created_date` - Date and time when the API Key was created.
* `last_updated_date` - Date and time when the API Key was last updated.
* `customer_id` - Amazon Web Services Marketplace customer identifier, when integrating with the Amazon Web Services SaaS Marketplace.
* `description` - Description of the API Key.
* `enabled` - Whether the API Key is enabled.
* `tags` - Map of tags for the resource.
