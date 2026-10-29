---
subcategory: "Bedrock Data Automation"
layout: "aws"
page_title: "AWS: aws_bedrockdataautomation_project"
description: |-
  Manages an Amazon Bedrock Data Automation Project.
---

# Resource: aws_bedrockdataautomation_project

Manages an Amazon Bedrock Data Automation Project. A project defines the standard output that Amazon Bedrock Data Automation extracts from files and the blueprints it uses for custom output.

~> **NOTE:** Amazon Bedrock Data Automation fills in default settings for every modality of `standard_output_configuration` and `override_configuration`. Terraform manages only the modalities you configure. Within a configured modality, configure every block the service would otherwise fill in, or Terraform shows the service defaults as differences. Changes made outside Terraform to modalities that aren't configured aren't detected.

## Example Usage

### Basic Usage

```terraform
resource "aws_bedrockdataautomation_project" "example" {
  name = "example"
}
```

### Standard and Custom Output

```terraform
resource "aws_bedrockdataautomation_blueprint" "example" {
  ### Configuration omitted for brevity ###
}

resource "aws_bedrockdataautomation_project" "example" {
  name        = "example"
  description = "Pay stub processing"

  standard_output_configuration {
    document {
      extraction {
        bounding_box {
          state = "ENABLED"
        }

        granularity {
          types = ["PAGE", "ELEMENT"]
        }
      }

      generative_field {
        state = "ENABLED"
      }

      output_format {
        additional_file_format {
          state = "DISABLED"
        }

        text_format {
          types = ["MARKDOWN"]
        }
      }
    }
  }

  custom_output_configuration {
    blueprint {
      blueprint_arn   = aws_bedrockdataautomation_blueprint.example.arn
      blueprint_stage = "LIVE"
    }
  }

  override_configuration {
    document {
      splitter {
        state = "ENABLED"
      }
    }
  }
}
```

## Argument Reference

The following arguments are required:

* `name` - (Required) Name of the project. Changing this value forces a new resource.

The following arguments are optional:

* `custom_output_configuration` - (Optional) Blueprints used for custom output. See [`custom_output_configuration` Block](#custom_output_configuration-block) below.
* `data_automation_library_configuration` - (Optional) Data automation libraries used by the project. See [`data_automation_library_configuration` Block](#data_automation_library_configuration-block) below.
* `description` - (Optional) Description of the project.
* `encryption_configuration` - (Optional) Encryption settings for the project. See [`encryption_configuration` Block](#encryption_configuration-block) below.
* `override_configuration` - (Optional) Processing overrides for each modality. See [`override_configuration` Block](#override_configuration-block) below.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `standard_output_configuration` - (Optional) Standard output to extract for each modality. See [`standard_output_configuration` Block](#standard_output_configuration-block) below.
* `tags` - (Optional) Map of tags assigned to the resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.
* `type` - (Optional) Type of API the project works with. Valid values: `ASYNC`, `SYNC`. Defaults to `ASYNC`. Changing this value forces a new resource.

### `custom_output_configuration` Block

* `blueprint` - (Optional) Blueprint to use for custom output. Can be specified multiple times. See [`blueprint` and `fallback_blueprint` Block](#blueprint-and-fallback_blueprint-block) below.
* `document` - (Optional) Custom output settings for documents. See [`custom_output_configuration.document` Block](#custom_output_configurationdocument-block) below.

### `custom_output_configuration.document` Block

* `fallback_blueprint` - (Optional) Blueprint to use when no other blueprint matches a document. Can be specified multiple times. See [`blueprint` and `fallback_blueprint` Block](#blueprint-and-fallback_blueprint-block) below.

### `blueprint` and `fallback_blueprint` Block

* `blueprint_arn` - (Required) ARN of the blueprint.
* `blueprint_stage` - (Optional) Stage of the blueprint. Valid values: `DEVELOPMENT`, `LIVE`.
* `blueprint_version` - (Optional) Version of the blueprint.

### `data_automation_library_configuration` Block

* `library` - (Optional) Library to use. Can be specified multiple times. See [`library` Block](#library-block) below.

### `library` Block

* `library_arn` - (Required) ARN of the data automation library.

### `encryption_configuration` Block

* `kms_encryption_context` - (Optional) Map of key-value pairs to use as the KMS encryption context.
* `kms_key_id` - (Required) Identifier of the KMS key used to encrypt the project.

### `override_configuration` Block

* `audio` - (Optional) Overrides for audio files. See [`override_configuration.audio` Block](#override_configurationaudio-block) below.
* `document` - (Optional) Overrides for documents. See [`override_configuration.document` Block](#override_configurationdocument-block) below.
* `image` - (Optional) Overrides for images. See [`override_configuration.image` Block](#override_configurationimage-block) below.
* `modality_routing` - (Optional) Modality to process each file type as. See [`modality_routing` Block](#modality_routing-block) below.
* `video` - (Optional) Overrides for videos. See [`override_configuration.video` Block](#override_configurationvideo-block) below.

### `override_configuration.audio` Block

* `language_configuration` - (Optional) Language settings. See [`language_configuration` Block](#language_configuration-block) below.
* `modality_processing` - (Optional) Whether to process audio files. See [`modality_processing` Block](#modality_processing-block) below.
* `sensitive_data_configuration` - (Optional) Sensitive data detection settings. See [`sensitive_data_configuration` Block](#sensitive_data_configuration-block) below.

### `language_configuration` Block

* `generative_output_language` - (Optional) Language of generative output. Valid values: `DEFAULT`, `EN`.
* `identify_multiple_languages` - (Optional) Whether to identify multiple languages in a file.
* `input_languages` - (Optional) Set of languages expected in input files. Valid values: `EN`, `DE`, `ES`, `FR`, `IT`, `PT`, `JA`, `KO`, `CN`, `TW`, `HK`.

### `override_configuration.document` Block

* `modality_processing` - (Optional) Whether to process documents. See [`modality_processing` Block](#modality_processing-block) below.
* `sensitive_data_configuration` - (Optional) Sensitive data detection settings. See [`sensitive_data_configuration` Block](#sensitive_data_configuration-block) below.
* `splitter` - (Optional) Whether to split documents that contain multiple logical documents. See [`splitter` Block](#splitter-block) below.

### `override_configuration.image` Block

* `modality_processing` - (Optional) Whether to process image files. See [`modality_processing` Block](#modality_processing-block) below.
* `sensitive_data_configuration` - (Optional) Sensitive data detection settings. See [`sensitive_data_configuration` Block](#sensitive_data_configuration-block) below.

### `override_configuration.video` Block

* `modality_processing` - (Optional) Whether to process video files. See [`modality_processing` Block](#modality_processing-block) below.
* `sensitive_data_configuration` - (Optional) Sensitive data detection settings. See [`sensitive_data_configuration` Block](#sensitive_data_configuration-block) below.

### `modality_routing` Block

Valid values for each argument: `IMAGE`, `DOCUMENT`, `AUDIO`, `VIDEO`.

* `jpeg` - (Optional) Modality to process JPEG files as.
* `mov` - (Optional) Modality to process MOV files as.
* `mp4` - (Optional) Modality to process MP4 files as.
* `png` - (Optional) Modality to process PNG files as.

### `modality_processing` Block

* `state` - (Required) Whether to process files of this modality. Valid values: `ENABLED`, `DISABLED`.

### `sensitive_data_configuration` Block

* `detection_mode` - (Required) Whether to detect or detect and redact sensitive data. Valid values: `DETECTION`, `DETECTION_AND_REDACTION`.
* `detection_scope` - (Optional) Set of outputs to scan for sensitive data. Valid values: `STANDARD`, `CUSTOM`.
* `pii_entities_configuration` - (Optional) PII entity settings. See [`pii_entities_configuration` Block](#pii_entities_configuration-block) below.

### `pii_entities_configuration` Block

* `pii_entity_types` - (Optional) Set of PII entity types to detect, such as `ALL`, `NAME`, `EMAIL` or `US_SOCIAL_SECURITY_NUMBER`. See the [API reference](https://docs.aws.amazon.com/bedrock/latest/APIReference/API_data-automation_PIIEntitiesConfiguration.html) for all valid values.
* `redaction_mask_mode` - (Optional) How to mask redacted data. Valid values: `PII`, `ENTITY_TYPE`.

### `splitter` Block

* `state` - (Required) Whether to split documents that contain multiple logical documents. Valid values: `ENABLED`, `DISABLED`.

### `standard_output_configuration` Block

* `audio` - (Optional) Standard output for audio files. See [`standard_output_configuration.audio` Block](#standard_output_configurationaudio-block) below.
* `document` - (Optional) Standard output for documents. See [`standard_output_configuration.document` Block](#standard_output_configurationdocument-block) below.
* `image` - (Optional) Standard output for images. See [`standard_output_configuration.image` Block](#standard_output_configurationimage-block) below.
* `video` - (Optional) Standard output for videos. See [`standard_output_configuration.video` Block](#standard_output_configurationvideo-block) below.

### `standard_output_configuration.audio` Block

* `extraction` - (Optional) Extraction settings. See [`standard_output_configuration.audio.extraction` Block](#standard_output_configurationaudioextraction-block) below.
* `generative_field` - (Optional) Generative output settings. See [`audio.generative_field` Block](#audiogenerative_field-block) below.

### `standard_output_configuration.audio.extraction` Block

* `category` - (Required) Categories to extract. See [`audio.category` Block](#audiocategory-block) below.

### `audio.category` Block

* `state` - (Required) Whether the setting is enabled. Valid values: `ENABLED`, `DISABLED`.
* `type_configuration` - (Optional) Per-category type settings. See [`type_configuration` Block](#type_configuration-block) below.
* `types` - (Optional) Set of categories to extract. Valid values: `AUDIO_CONTENT_MODERATION`, `TRANSCRIPT`, `TOPIC_CONTENT_MODERATION`.

### `type_configuration` Block

* `transcript` - (Optional) Transcript settings. See [`transcript` Block](#transcript-block) below.

### `transcript` Block

* `channel_labeling` - (Optional) Whether to label audio channels. See [`channel_labeling` Block](#channel_labeling-block) below.
* `speaker_labeling` - (Optional) Whether to label speakers. See [`speaker_labeling` Block](#speaker_labeling-block) below.

### `channel_labeling` Block

* `state` - (Required) Whether to label audio channels. Valid values: `ENABLED`, `DISABLED`.

### `speaker_labeling` Block

* `state` - (Required) Whether to label speakers. Valid values: `ENABLED`, `DISABLED`.

### `audio.generative_field` Block

* `state` - (Required) Whether the setting is enabled. Valid values: `ENABLED`, `DISABLED`.
* `types` - (Optional) Set of types to produce. Valid values: `AUDIO_SUMMARY`, `IAB`, `TOPIC_SUMMARY`.

### `standard_output_configuration.document` Block

* `extraction` - (Optional) Extraction settings. See [`standard_output_configuration.document.extraction` Block](#standard_output_configurationdocumentextraction-block) below.
* `generative_field` - (Optional) Whether to generate document summaries. See [`document.generative_field` Block](#documentgenerative_field-block) below.
* `output_format` - (Optional) Output format settings. See [`output_format` Block](#output_format-block) below.

### `standard_output_configuration.document.extraction` Block

* `bounding_box` - (Required) Whether to include bounding boxes. See [`bounding_box` Block](#bounding_box-block) below.
* `granularity` - (Required) Extraction granularity. See [`granularity` Block](#granularity-block) below.

### `bounding_box` Block

* `state` - (Required) Whether to include bounding boxes. Valid values: `ENABLED`, `DISABLED`.

### `granularity` Block

* `types` - (Optional) Set of granularity levels. Valid values: `DOCUMENT`, `PAGE`, `ELEMENT`, `WORD`, `LINE`.

### `document.generative_field` Block

* `state` - (Required) Whether to generate document summaries. Valid values: `ENABLED`, `DISABLED`.

### `output_format` Block

* `additional_file_format` - (Required) Whether to produce additional output files. See [`additional_file_format` Block](#additional_file_format-block) below.
* `text_format` - (Required) Text output formats. See [`text_format` Block](#text_format-block) below.

### `additional_file_format` Block

* `state` - (Required) Whether to produce additional output files. Valid values: `ENABLED`, `DISABLED`.

### `text_format` Block

* `types` - (Optional) Set of text formats. Valid values: `PLAIN_TEXT`, `MARKDOWN`, `HTML`, `CSV`.

### `standard_output_configuration.image` Block

* `extraction` - (Optional) Extraction settings. See [`standard_output_configuration.image.extraction` Block](#standard_output_configurationimageextraction-block) below.
* `generative_field` - (Optional) Generative output settings. See [`image.generative_field` Block](#imagegenerative_field-block) below.

### `standard_output_configuration.image.extraction` Block

* `bounding_box` - (Required) Whether to include bounding boxes. See [`bounding_box` Block](#bounding_box-block) below.
* `category` - (Required) Categories to extract. See [`image.category` Block](#imagecategory-block) below.

### `image.category` Block

* `state` - (Required) Whether the setting is enabled. Valid values: `ENABLED`, `DISABLED`.
* `types` - (Optional) Set of categories to extract. Valid values: `CONTENT_MODERATION`, `TEXT_DETECTION`, `LOGOS`.

### `image.generative_field` Block

* `state` - (Required) Whether the setting is enabled. Valid values: `ENABLED`, `DISABLED`.
* `types` - (Optional) Set of types to produce. Valid values: `IMAGE_SUMMARY`, `IAB`.

### `standard_output_configuration.video` Block

* `extraction` - (Optional) Extraction settings. See [`standard_output_configuration.video.extraction` Block](#standard_output_configurationvideoextraction-block) below.
* `generative_field` - (Optional) Generative output settings. See [`video.generative_field` Block](#videogenerative_field-block) below.

### `standard_output_configuration.video.extraction` Block

* `bounding_box` - (Required) Whether to include bounding boxes. See [`bounding_box` Block](#bounding_box-block) below.
* `category` - (Required) Categories to extract. See [`video.category` Block](#videocategory-block) below.

### `video.category` Block

* `state` - (Required) Whether the setting is enabled. Valid values: `ENABLED`, `DISABLED`.
* `types` - (Optional) Set of categories to extract. Valid values: `CONTENT_MODERATION`, `TEXT_DETECTION`, `TRANSCRIPT`, `LOGOS`.

### `video.generative_field` Block

* `state` - (Required) Whether the setting is enabled. Valid values: `ENABLED`, `DISABLED`.
* `types` - (Optional) Set of types to produce. Valid values: `VIDEO_SUMMARY`, `IAB`, `CHAPTER_SUMMARY`.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the project.
* `creation_time` - Time the project was created, in RFC3339 format.
* `last_modified_time` - Time the project was last modified, in RFC3339 format.
* `project_stage` - Stage of the project.
* `status` - Status of the project.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `10m`)
* `update` - (Default `10m`)
* `delete` - (Default `10m`)

## Import

Import doesn't populate `standard_output_configuration` or `override_configuration`. The next plan shows them as updates, and applying the plan sends the configured values to the service.

In Terraform v1.12.0 and later, the [`import` block](https://developer.hashicorp.com/terraform/language/import) can be used with the `identity` attribute. For example:

```terraform
import {
  to = aws_bedrockdataautomation_project.example
  identity = {
    arn = "arn:aws:bedrock:us-west-2:123456789012:data-automation-project/abcd1234efgh"
  }
}

resource "aws_bedrockdataautomation_project" "example" {
  ### Configuration omitted for brevity ###
}
```

### Identity Schema

#### Required

* `arn` (String) ARN of the project.

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Bedrock Data Automation Project using the `arn`. For example:

```terraform
import {
  to = aws_bedrockdataautomation_project.example
  id = "arn:aws:bedrock:us-west-2:123456789012:data-automation-project/abcd1234efgh"
}
```

Using `terraform import`, import Bedrock Data Automation Project using the `arn`. For example:

```console
% terraform import aws_bedrockdataautomation_project.example arn:aws:bedrock:us-west-2:123456789012:data-automation-project/abcd1234efgh
```
