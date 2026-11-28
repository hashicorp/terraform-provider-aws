---
subcategory: "SSM (Systems Manager)"
layout: "aws"
page_title: "AWS: aws_ssm_parameter"
description: |-
  Retrieve information about an SSM parameter, including its value. To retrieve parameter metadata, see the `aws_ssm_parameter` data source.
---

# Ephemeral: aws_ssm_parameter

Retrieve information about an SSM parameter, including its value.

~> **NOTE:** Ephemeral resources are a new feature and may evolve as we continue to explore their most effective uses. [Learn more](https://developer.hashicorp.com/terraform/language/resources/ephemeral).

## Example Usage

### Retrieve an SSM parameter

By default, this ephemeral resource attempts to return decrypted values for secure string parameters.

```terraform
ephemeral "aws_ssm_parameter" "example" {
  arn = aws_ssm_parameter.example.arn
}
```

## Argument Reference

This ephemeral resource supports the following arguments:

* `account` - (Optional) AWS account ID where this ephemeral resource will be opened. Defaults to the account of the provider's credentials. See the [Per-Resource Account guide](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/guides/per-resource-account).
* `arn` - (Required) ARN of the parameter that you want to query
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `with_decryption` - (Optional) Return decrypted values for a secure string parameter (Defaults to `true`).

## Attribute Reference

This ephemeral resource exports the following attributes in addition to the arguments above:

* `name` - Name of the parameter.
* `type` - Type of parameter.
* `value` - Parameter value.
* `version` - Parameter version.
