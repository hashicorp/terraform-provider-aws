---
subcategory: "Bedrock"
layout: "aws"
page_title: "AWS: aws_bedrock_guardrail_version"
description: |-
  Terraform resource for managing an AWS Bedrock Guardrail Version.
---
# Resource: aws_bedrock_guardrail_version

Terraform resource for managing an AWS Bedrock Guardrail Version.

## Example Usage

### Basic Usage

```terraform
resource "aws_bedrock_guardrail_version" "example" {
  description   = "example"
  guardrail_arn = aws_bedrock_guardrail.test.guardrail_arn
  skip_destroy  = true
}
```

## Argument Reference

The following arguments are required:

* `guardrail_arn` - (Required) Guardrail ARN.

The following arguments are optional:

* `description` - (Optional) Description of the Guardrail version.
* `skip_destroy` - (Optional) Whether to retain the old version of a previously deployed Guardrail. Default is `false`

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `version` - Guardrail version.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `5m`)
* `delete` - (Default `5m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Amazon Bedrock Guardrail Version using a comma-delimited string of `guardrail_arn` and `version`. For example:

```terraform
import {
  to = aws_bedrock_guardrail_version.example
  id = "arn:aws:bedrock:us-west-2:123456789012:guardrail-id-12345678,1"
}
```

Using `terraform import`, import Amazon Bedrock Guardrail Version using using a comma-delimited string of `guardrail_arn` and `version`. For example:

```console
% terraform import aws_bedrock_guardrail_version.example arn:aws:bedrock:us-west-2:123456789012:guardrail-id-12345678,1
```
