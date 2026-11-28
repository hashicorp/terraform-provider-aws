---
subcategory: "API Gateway"
layout: "aws"
page_title: "AWS: aws_api_gateway_api_key"
description: |-
  Get information on an API Gateway REST API Key
---

# Data Source: aws_api_gateway_api_key

Use this data source to get the name and value of a pre-existing API Key, for
example to supply credentials for a dependency microservice.

## Example Usage

```terraform
data "aws_api_gateway_api_key" "my_api_key" {
  id = "ru3mpjgse6"
}
```

## Argument Reference

This data source supports the following arguments:

* `account` - (Optional) AWS account ID where this data source will be read. Defaults to the account of the provider's credentials. See the [Per-Resource Account guide](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/guides/per-resource-account).
* `id` - (Required) ID of the API Key to look up.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `arn` - ARN of the API Key.
* `created_date` - Date and time when the API Key was created.
* `customer_id` - Amazon Web Services Marketplace customer identifier, when integrating with the Amazon Web Services SaaS Marketplace.
* `description` - Description of the API Key.
* `enabled` - Whether the API Key is enabled.
* `last_updated_date` - Date and time when the API Key was last updated.
* `name` - Set to the name of the API Key.
* `tags` - Map of tags for the resource.
* `value` - Set to the value of the API Key.
