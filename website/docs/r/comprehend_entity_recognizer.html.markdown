---
subcategory: "Comprehend"
layout: "aws"
page_title: "AWS: aws_comprehend_entity_recognizer"
description: |-
  Terraform resource for managing an AWS Comprehend Entity Recognizer.
---

# Resource: aws_comprehend_entity_recognizer

Terraform resource for managing an AWS Comprehend Entity Recognizer.

## Example Usage

### Basic Usage

```terraform
resource "aws_comprehend_entity_recognizer" "example" {
  name = "example"

  data_access_role_arn = aws_iam_role.example.arn

  language_code = "en"
  input_data_config {
    entity_types {
      type = "ENTITY_1"
    }
    entity_types {
      type = "ENTITY_2"
    }

    documents {
      s3_uri = "s3://${aws_s3_bucket.documents.bucket}/${aws_s3_object.documents.id}"
    }

    entity_list {
      s3_uri = "s3://${aws_s3_bucket.entities.bucket}/${aws_s3_object.entities.id}"
    }
  }

  depends_on = [
    aws_iam_role_policy.example
  ]
}

resource "aws_s3_object" "documents" {
  # ...
}

resource "aws_s3_object" "entities" {
  # ...
}
```

## Argument Reference

The following arguments are required:

* `data_access_role_arn` - (Required) The ARN for an IAM Role which allows Comprehend to read the training and testing data.
* `input_data_config` - (Required) Configuration for the training and testing data.
  See the [`input_data_config` Configuration Block](#input_data_config-configuration-block) section below.
* `language_code` - (Required) Two-letter language code for the language.
  One of `en`, `es`, `fr`, `it`, `de`, or `pt`.
* `name` - (Required) Name for the Entity Recognizer.
  Has a maximum length of 63 characters.
  Can contain upper- and lower-case letters, numbers, and hypen (`-`).

The following arguments are optional:

* `model_kms_key_id` - (Optional) The ID or ARN of a KMS Key used to encrypt trained Entity Recognizers.
* `tags` - (Optional) A map of tags to assign to the resource. If configured with a provider [`default_tags` Configuration Block](/docs/providers/aws/index.html#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.
* `version_name` - (Optional) Name for the version of the Entity Recognizer.
  Each version must have a unique name within the Entity Recognizer.
  If omitted, Terraform will assign a random, unique version name.
  If explicitly set to `""`, no version name will be set.
  Has a maximum length of 63 characters.
  Can contain upper- and lower-case letters, numbers, and hypen (`-`).
  Conflicts with `version_name_prefix`.
* `version_name_prefix` - (Optional) Creates a unique version name beginning with the specified prefix.
  Has a maximum length of 37 characters.
  Can contain upper- and lower-case letters, numbers, and hypen (`-`).
  Conflicts with `version_name`.
* `volume_kms_key_id` - (Optional) ID or ARN of a KMS Key used to encrypt storage volumes during job processing.
* `vpc_config` - (Optional) Configuration parameters for VPC to contain Entity Recognizer resources.
  See the [`vpc_config` Configuration Block](#vpc_config-configuration-block) section below.

### `input_data_config` Configuration Block

* `annotations` - (Optional) Specifies location of the document annotation data.
  See the [`annotations` Configuration Block](#annotations-configuration-block) section below.
  One of `annotations` or `entity_list` is required.
* `augmented_manifests` - (Optional) List of training datasets produced by Amazon SageMaker Ground Truth.
  Used if `data_format` is `AUGMENTED_MANIFEST`.
  See the [`augmented_manifests` Configuration Block](#augmented_manifests-configuration-block) section below.
* `data_format` - (Optional, Default: `COMPREHEND_CSV`) The format for the training data.
  One of `COMPREHEND_CSV` or `AUGMENTED_MANIFEST`.
* `documents` - (Optional) Specifies a collection of training documents.
  Used if `data_format` is `COMPREHEND_CSV`.
  See the [`documents` Configuration Block](#documents-configuration-block) section below.
* `entity_list` - (Optional) Specifies location of the entity list data.
  See the [`entity_list` Configuration Block](#entity_list-configuration-block) section below.
  One of `entity_list` or `annotations` is required.
* `entity_types` - (Required) Set of entity types to be recognized.
  Has a maximum of 25 items.
  See the [`entity_types` Configuration Block](#entity_types-configuration-block) section below.

### `annotations` Configuration Block

* `s3_uri` - (Required) Location of training annotations.
* `test_s3uri` - (Optional) Location of test annotations.

### `augmented_manifests` Configuration Block

* `annotation_data_s3_uri` - (Optional) Location of annotation files.
* `attribute_names` - (Required) The JSON attribute that contains the annotations for the training documents.
* `document_type` - (Optional, Default: `PLAIN_TEXT_DOCUMENT`) Type of augmented manifest.
  One of `PLAIN_TEXT_DOCUMENT` or `SEMI_STRUCTURED_DOCUMENT`.
* `s3_uri` - (Required) Location of augmented manifest file.
* `source_documents_s3_uri` - (Optional) Location of source PDF files.
* `split` - (Optional, Default: `TRAIN`) Purpose of data in augmented manifest.
  One of `TRAIN` or `TEST`.

### `documents` Configuration Block

* `input_format` - (Optional, Default: `ONE_DOC_PER_LINE`) Specifies how the input files should be processed.
  One of `ONE_DOC_PER_LINE` or `ONE_DOC_PER_FILE`.
* `s3_uri` - (Required) Location of training documents.
* `test_s3uri` - (Optional) Location of test documents.

### `entity_list` Configuration Block

* `s3_uri` - (Required) Location of entity list.

### `entity_types` Configuration Block

* `type` - (Required) An entity type to be matched by the Entity Recognizer.
  Cannot contain a newline (`\n`), carriage return (`\r`), or tab (`\t`).

### `vpc_config` Configuration Block

* `security_group_ids` - (Required) List of security group IDs.
* `subnets` - (Required) List of VPC subnets.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the Entity Recognizer version.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block).

## Timeouts

`aws_comprehend_entity_recognizer` provides the following [Timeouts](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts) configuration options:

* `create` - (Optional, Default: `60m`)
* `update` - (Optional, Default: `60m`)
* `delete` - (Optional, Default: `30m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Comprehend Entity Recognizer using the ARN. For example:

```terraform
import {
  to = aws_comprehend_entity_recognizer.example
  id = "arn:aws:comprehend:us-west-2:123456789012:entity-recognizer/example"
}
```

Using `terraform import`, import Comprehend Entity Recognizer using the ARN. For example:

```console
% terraform import aws_comprehend_entity_recognizer.example arn:aws:comprehend:us-west-2:123456789012:entity-recognizer/example
```
