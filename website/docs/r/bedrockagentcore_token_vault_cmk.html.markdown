---
subcategory: "Bedrock AgentCore"
layout: "aws"
page_title: "AWS: aws_bedrockagentcore_token_vault_cmk"
description: |-
  Manages the customer master key (CMK) for a token vault.
---

# Resource: aws_bedrockagentcore_token_vault_cmk

Manages the AWS KMS customer master key (CMK) for a token vault.

~> Deletion of this resource will not modify the CMK, only remove the resource from state.

## Example Usage

```terraform
resource "aws_bedrockagentcore_token_vault_cmk" "example" {
  kms_configuration {
    key_type    = "CustomerManagedKey"
    kms_key_arn = aws_kms_key.example.arn
  }
}
```

## Argument Reference

This resource supports the following arguments:

* `kms_configuration` - (Required) KMS configuration for the token vault. See [`kms_configuration`](#kms_configuration) below.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `token_vault_id` - (Optional) Token vault ID. Defaults to `default`.

### `kms_configuration`

The `kms_configuration` block supports the following:

* `key_type` - (Required) Type of KMS key. Valid values: `CustomerManagedKey`, `ServiceManagedKey`.
* `kms_key_arn` - (Optional) ARN of the KMS key.

## Attribute Reference

This resource exports no additional attributes.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import token vault CMKs using the token vault ID. For example:

```terraform
import {
  to = aws_bedrockagentcore_token_vault_cmk.example
  id = "default"
}
```

Using `terraform import`, import token vault CMKs using the token vault ID. For example:

```console
% terraform import aws_bedrockagentcore_token_vault_cmk.example "default"
```
