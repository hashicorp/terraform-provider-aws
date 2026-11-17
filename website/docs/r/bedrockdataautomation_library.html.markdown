---
subcategory: "Bedrock Data Automation"
layout: "aws"
page_title: "AWS: aws_bedrockdataautomation_library"
description: |-
  Manages an Amazon Bedrock Data Automation Library.
---

# Resource: aws_bedrockdataautomation_library

Manages an Amazon Bedrock Data Automation Library. A library holds entities, such as custom vocabulary, that a project uses when it processes files.

## Example Usage

### Basic Usage

```terraform
resource "aws_bedrockdataautomation_library" "example" {
  name        = "example"
  description = "Medical vocabulary"
}
```

### Use With a Project

```terraform
resource "aws_bedrockdataautomation_library" "example" {
  name = "example"
}

resource "aws_bedrockdataautomation_project" "example" {
  name = "example"

  data_automation_library_configuration {
    library {
      library_arn = aws_bedrockdataautomation_library.example.arn
    }
  }
}
```

## Argument Reference

The following arguments are required:

* `name` - (Required) Name of the library. Changing this value forces a new resource.

The following arguments are optional:

* `description` - (Optional) Description of the library. Removing the description forces a new resource, because the service can't remove a description from an existing library.
* `encryption_configuration` - (Optional) Encryption settings for the library. Changing this block forces a new resource. See [`encryption_configuration` Block](#encryption_configuration-block) below.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `tags` - (Optional) Map of tags assigned to the resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

### `encryption_configuration` Block

* `kms_encryption_context` - (Optional) Map of key-value pairs to use as the KMS encryption context.
* `kms_key_id` - (Required) Identifier of the KMS key used to encrypt the library.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the library.
* `creation_time` - Time the library was created, in RFC3339 format.
* `entity_types` - Entity types in the library. See [`entity_types` Block](#entity_types-block) below.
* `status` - Status of the library.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

### `entity_types` Block

* `entity_metadata` - JSON metadata for the entity type.
* `entity_type` - Entity type, such as `VOCABULARY`.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `delete` - (Default `10m`)

## Import

In Terraform v1.12.0 and later, the [`import` block](https://developer.hashicorp.com/terraform/language/import) can be used with the `identity` attribute. For example:

```terraform
import {
  to = aws_bedrockdataautomation_library.example
  identity = {
    arn = "arn:aws:bedrock:us-west-2:123456789012:data-automation-library/abcd1234efgh"
  }
}

resource "aws_bedrockdataautomation_library" "example" {
  ### Configuration omitted for brevity ###
}
```

### Identity Schema

#### Required

* `arn` (String) ARN of the library.

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Bedrock Data Automation Library using the `arn`. For example:

```terraform
import {
  to = aws_bedrockdataautomation_library.example
  id = "arn:aws:bedrock:us-west-2:123456789012:data-automation-library/abcd1234efgh"
}
```

Using `terraform import`, import Bedrock Data Automation Library using the `arn`. For example:

```console
% terraform import aws_bedrockdataautomation_library.example arn:aws:bedrock:us-west-2:123456789012:data-automation-library/abcd1234efgh
```
