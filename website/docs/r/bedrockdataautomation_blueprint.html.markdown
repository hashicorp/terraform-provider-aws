---
subcategory: "Bedrock Data Automation"
layout: "aws"
page_title: "AWS: aws_bedrockdataautomation_blueprint"
description: |-
  Manages an Amazon Bedrock Data Automation Blueprint.
---

# Resource: aws_bedrockdataautomation_blueprint

Manages an Amazon Bedrock Data Automation Blueprint. A blueprint defines the fields to extract from a document, image, audio, or video file.

## Example Usage

### Basic Usage

```terraform
resource "aws_bedrockdataautomation_blueprint" "example" {
  name = "example"
  type = "DOCUMENT"

  schema = jsonencode({
    "$schema"   = "http://json-schema.org/draft-07/schema#"
    description = "Pay stub"
    class       = "Pay stub"
    type        = "object"
    definitions = {}
    properties = {
      gross_pay = {
        type          = "number"
        inferenceType = "explicit"
        instruction   = "The gross pay for this pay period"
      }
      net_pay = {
        type          = "number"
        inferenceType = "explicit"
        instruction   = "The net pay for this pay period"
      }
    }
  })
}
```

### With Encryption

```terraform
resource "aws_bedrockdataautomation_blueprint" "example" {
  name   = "example"
  type   = "DOCUMENT"
  schema = file("${path.module}/blueprint.json")

  encryption_configuration {
    kms_key_id = aws_kms_key.example.arn

    kms_encryption_context = {
      Department = "payroll"
    }
  }
}
```

## Argument Reference

The following arguments are required:

* `name` - (Required) Name of the blueprint. Changing this value forces a new resource.
* `schema` - (Required) JSON schema of the blueprint, defining the fields to extract.
* `type` - (Required) Modality of the blueprint. Valid values: `DOCUMENT`, `IMAGE`, `AUDIO`, `VIDEO`. Changing this value forces a new resource.

The following arguments are optional:

* `encryption_configuration` - (Optional) Encryption settings for the blueprint. See [`encryption_configuration`](#encryption_configuration-block) below.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `tags` - (Optional) Map of tags assigned to the resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

### `encryption_configuration` Block

* `kms_encryption_context` - (Optional) Map of key-value pairs to use as the KMS encryption context.
* `kms_key_id` - (Required) Identifier of the KMS key used to encrypt the blueprint.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the blueprint.
* `blueprint_stage` - Stage of the blueprint.
* `creation_time` - Time the blueprint was created, in RFC3339 format.
* `last_modified_time` - Time the blueprint was last modified, in RFC3339 format.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Import

In Terraform v1.12.0 and later, the [`import` block](https://developer.hashicorp.com/terraform/language/import) can be used with the `identity` attribute. For example:

```terraform
import {
  to = aws_bedrockdataautomation_blueprint.example
  identity = {
    arn = "arn:aws:bedrock:us-west-2:123456789012:blueprint/abcd1234efgh"
  }
}

resource "aws_bedrockdataautomation_blueprint" "example" {
  ### Configuration omitted for brevity ###
}
```

### Identity Schema

#### Required

* `arn` (String) ARN of the blueprint.

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Bedrock Data Automation Blueprint using the `arn`. For example:

```terraform
import {
  to = aws_bedrockdataautomation_blueprint.example
  id = "arn:aws:bedrock:us-west-2:123456789012:blueprint/abcd1234efgh"
}
```

Using `terraform import`, import Bedrock Data Automation Blueprint using the `arn`. For example:

```console
% terraform import aws_bedrockdataautomation_blueprint.example arn:aws:bedrock:us-west-2:123456789012:blueprint/abcd1234efgh
```
