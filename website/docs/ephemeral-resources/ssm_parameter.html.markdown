---
subcategory: "SSM (Systems Manager)"
layout: "aws"
page_title: "AWS: aws_ssm_parameter"
description: |-
  Retrieve information about an SSM parameter, including its value. To retrieve parameter metadata, see the `aws_ssm_parameter` data source.
---

# Ephemeral: aws_ssm_parameter

Retrieve information about an SSM parameter, including its value.

~> **NOTE:** Ephemeral resources are a new feature and may evolve as we continue to explore their most effective uses. [Learn more](https://developer.hashicorp.com/terraform/language/v1.10.x/resources/ephemeral).

## Example Usage

### Retrieve an SSM parameter

By default, this ephemeral resource attempts to return decrypted values for secure string parameters.

```terraform
ephemeral "aws_ssm_parameter" "example" {
  arn = aws_ssm_parameter.example.arn
}
```

## Argument Reference

* `arn` - (Required) The Amazon Resource Name (ARN) of the parameter that you want to query
* `with_decryption` - (Optional) Return decrypted values for a secure string parameter (Defaults to `true`).

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `name` - The name of the parameter.
* `type` - The type of parameter.
* `value` - The parameter value.
* `version` - The parameter version.
* `with_decryption` - Indicates whether the secure string parameters were decrypted.
