---
subcategory: "API Gateway"
layout: "aws"
page_title: "AWS: aws_api_gateway_request_validator"
description: |-
  Manages an API Gateway Request Validator.
---

# Resource: aws_api_gateway_request_validator

Manages an API Gateway Request Validator.

## Example Usage

```terraform
resource "aws_api_gateway_request_validator" "example" {
  name                        = "example"
  rest_api_id                 = aws_api_gateway_rest_api.example.id
  validate_request_body       = true
  validate_request_parameters = true
}
```

## Argument Reference

This resource supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `name` - (Required) Name of the request validator
* `rest_api_id` - (Required) ID of the associated Rest API
* `validate_request_body` - (Optional) Boolean whether to validate request body. Defaults to `false`.
* `validate_request_parameters` - (Optional) Boolean whether to validate request parameters. Defaults to `false`.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - Unique ID of the request validator

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import `aws_api_gateway_request_validator` using `REST-API-ID/REQUEST-VALIDATOR-ID`. For example:

```terraform
import {
  to = aws_api_gateway_request_validator.example
  id = "12345abcde/67890fghij"
}
```

Using `terraform import`, import `aws_api_gateway_request_validator` using `REST-API-ID/REQUEST-VALIDATOR-ID`. For example:

```console
% terraform import aws_api_gateway_request_validator.example 12345abcde/67890fghij
```
