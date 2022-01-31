---
subcategory: "IoT"
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
    Key1 = "${file("test-fixtures/iot-authorizer-signing-key.pem")}"
  }
}
```

## Argument Reference

* `authorizer_function_arn` - (Required) The ARN of the authorizer's Lambda function.
* `name` - (Required) The name of the authorizer.
* `signing_disabled` - (Optional) Specifies whether AWS IoT validates the token signature in an authorization request. Default: `false`.
* `status` - (Optional) The status of Authorizer request at creation. Valid values: `ACTIVE`, `INACTIVE`. Default: `ACTIVE`.
* `token_key_name` - (Optional) The name of the token key used to extract the token from the HTTP headers. This value is required if signing is enabled in your authorizer.
* `token_signing_public_keys` - (Optional) The public keys used to verify the digital signature returned by your custom authentication service. This value is required if signing is enabled in your authorizer.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - The ARN of the authorizer.

## Import

IOT Authorizers can be imported using the name, e.g.,

```
$ terraform import aws_iot_authorizer.example example
```
