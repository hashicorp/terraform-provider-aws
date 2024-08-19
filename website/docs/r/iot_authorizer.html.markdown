---
subcategory: "IoT Core"
layout: "aws"
page_title: "AWS: aws_iot_authorizer"
description: |-
    Creates and manages an AWS IoT Authorizer.
---

# Resource: aws_iot_authorizer

Creates and manages an AWS IoT Authorizer.

## Example Usage

```terraform
resource "aws_iot_authorizer" "example" {
  name                    = "example"
  authorizer_function_arn = aws_lambda_function.example.arn
  signing_disabled        = false
  status                  = "ACTIVE"
  token_key_name          = "Token-Header"

  token_signing_public_keys = {
    Key1 = file("test-fixtures/iot-authorizer-signing-key.pem")
  }

  tags = {
    Name = "example"
  }
}
```

## Argument Reference

* `authorizer_function_arn` - (Required) The ARN of the authorizer's Lambda function.
* `enable_caching_for_http`  - (Optional) Specifies whether the HTTP caching is enabled or not. Default: `false`.
* `name` - (Required) The name of the authorizer.
* `signing_disabled` - (Optional) Specifies whether AWS IoT validates the token signature in an authorization request. Default: `false`.
* `status` - (Optional) The status of Authorizer request at creation. Valid values: `ACTIVE`, `INACTIVE`. Default: `ACTIVE`.
* `tags` - (Optional) Map of tags to assign to this resource. If configured with a provider [`default_tags` configuration block](https://www.terraform.io/docs/providers/aws/index.html#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.
* `token_key_name` - (Optional) The name of the token key used to extract the token from the HTTP headers. This value is required if signing is enabled in your authorizer.
* `token_signing_public_keys` - (Optional) The public keys used to verify the digital signature returned by your custom authentication service. This value is required if signing is enabled in your authorizer.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - The ARN of the authorizer.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://www.terraform.io/docs/providers/aws/index.html#default_tags-configuration-block).

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import IOT Authorizers using the name. For example:

```terraform
import {
  to = aws_iot_authorizer.example
  id = "example"
}
```

Using `terraform import`, import IOT Authorizers using the name. For example:

```console
% terraform import aws_iot_authorizer.example example
```
