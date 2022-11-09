---
subcategory: "Comprehend"
layout: "aws"
page_title: "AWS: aws_comprehend_document_classifier"
description: |-
  Terraform resource for managing an AWS Comprehend Document Classifier.
---

# Resource: aws_comprehend_document_classifier

Terraform resource for managing an AWS Comprehend Document Classifier.

## Example Usage

### Basic Usage

```terraform
resource "aws_comprehend_document_classifier" "example" {
  name = "example"

  data_access_role_arn = aws_iam_role.example.arn

  language_code = "en"
  input_data_config {
    s3_uri = "s3://${aws_s3_bucket.test.bucket}/${aws_s3_object.documents.id}"
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
* `name` - (Required) Name for the Document Classifier.
  Has a maximum length of 63 characters.
  Can contain upper- and lower-case letters, numbers, and hypen (`-`).

The following arguments are optional:

* `mode` - (Optional, Default: `MULTI_CLASS`) The document classification mode.
  One of `MULTI_CLASS` or `MULTI_LABEL`.
  `MULTI_CLASS` is also known as "Single Label" in the AWS Console.
* `model_kms_key_id` - (Optional) KMS Key used to encrypt trained Document Classifiers.
  Can be a KMS Key ID or a KMS Key ARN.
* `output_data_config` - (Optional) Configuration for the output results of training.
  See the [`output_data_config` Configuration Block](#output_data_config-configuration-block) section below.
* `tags` - (Optional) A map of tags to assign to the resource. If configured with a provider [`default_tags` Configuration Block](/docs/providers/aws/index.html#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.
* `version_name` - (Optional) Name for the version of the Document Classifier.
  Each version must have a unique name within the Document Classifier.
  If omitted, Terraform will assign a random, unique version name.
  If explicitly set to `""`, no version name will be set.
  Has a maximum length of 63 characters.
  Can contain upper- and lower-case letters, numbers, and hypen (`-`).
  Conflicts with `version_name_prefix`.
* `version_name_prefix` - (Optional) Creates a unique version name beginning with the specified prefix.
  Has a maximum length of 37 characters.
  Can contain upper- and lower-case letters, numbers, and hypen (`-`).
  Conflicts with `version_name`.
* `volume_kms_key_id` - (Optional) KMS Key used to encrypt storage volumes during job processing.
  Can be a KMS Key ID or a KMS Key ARN.
* `vpc_config` - (Optional) Configuration parameters for VPC to contain Document Classifier resources.
  See the [`vpc_config` Configuration Block](#vpc_config-configuration-block) section below.

### `input_data_config` Configuration Block

* `augmented_manifests` - (Optional) List of training datasets produced by Amazon SageMaker Ground Truth.
  Used if `data_format` is `AUGMENTED_MANIFEST`.
  See the [`augmented_manifests` Configuration Block](#augmented_manifests-configuration-block) section below.
* `data_format` - (Optional, Default: `COMPREHEND_CSV`) The format for the training data.
  One of `COMPREHEND_CSV` or `AUGMENTED_MANIFEST`.
* `label_delimiter` - (Optional) Delimiter between labels when training a multi-label classifier.
  Valid values are `|`, `~`, `!`, `@`, `#`, `$`, `%`, `^`, `*`, `-`, `_`, `+`, `=`, `\`, `:`, `;`, `>`, `?`, `/`, `<space>`, and `<tab>`.
  Default is `|`.
* `s3_uri` - (Optional) Location of training documents.
  Used if `data_format` is `COMPREHEND_CSV`.
* `test_s3uri` - (Optional) Location of test documents.

### `augmented_manifests` Configuration Block

* `annotation_data_s3_uri` - (Optional) Location of annotation files.
* `attribute_names` - (Required) The JSON attribute that contains the annotations for the training documents.
* `document_type` - (Optional, Default: `PLAIN_TEXT_DOCUMENT`) Type of augmented manifest.
  One of `PLAIN_TEXT_DOCUMENT` or `SEMI_STRUCTURED_DOCUMENT`.
* `s3_uri` - (Required) Location of augmented manifest file.
* `source_documents_s3_uri` - (Optional) Location of source PDF files.
* `split` - (Optional, Default: `TRAIN`) Purpose of data in augmented manifest.
  One of `TRAIN` or `TEST`.

### `output_data_config` Configuration Block

* `kms_key_id` - (Optional) KMS Key used to encrypt the output documents.
  Can be a KMS Key ID, a KMS Key ARN, a KMS Alias name, or a KMS Alias ARN.
* `output_s3_uri` - (Computed) Full path for the output documents.
* `s3_uri` - (Required) Destination path for the output documents.
  The full path to the output file will be returned in `output_s3_uri`.

### `vpc_config` Configuration Block

* `security_group_ids` - (Required) List of security group IDs.
* `subnets` - (Required) List of VPC subnets.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - ARN of the Document Classifier version.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block).

## Timeouts

`aws_comprehend_document_classifier` provides the following [Timeouts](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts) configuration options:

* `create` - (Optional, Default: `60m`)
* `update` - (Optional, Default: `60m`)
* `delete` - (Optional, Default: `30m`)

## Import

Comprehend Document Classifier can be imported using the ARN, e.g.,

```
$ terraform import aws_comprehend_document_classifier.example arn:aws:comprehend:us-west-2:123456789012:document_classifier/example
```
