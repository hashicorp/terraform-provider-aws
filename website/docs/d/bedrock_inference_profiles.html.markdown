---
subcategory: "Bedrock"
layout: "aws"
page_title: "AWS: aws_bedrock_inference_profiles"
description: |-
  Terraform data source for managing AWS Bedrock Inference Profiles.
---

# Data Source: aws_bedrock_inference_profiles

Terraform data source for managing AWS Bedrock AWS Bedrock Inference Profiles.

## Example Usage

### Basic Usage

```terraform
data "aws_bedrock_inference_profiles" "test" {}
```

## Argument Reference

There are no arguments available for this data source.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

- `inference_profile_summaries` - List of inference profile summary objects. See [`inference_profile_summaries`](#inference_profile_summaries).

### `inference_profile_summaries`

- `created_at` - The time at which the inference profile was created.
- `description` - The description of the inference profile.
- `inference_profile_arn` - The Amazon Resource Name (ARN) of the inference profile.
- `inference_profile_id` - The unique identifier of the inference profile.
- `inference_profile_name` - The name of the inference profile.
- `models` - A list of information about each model in the inference profile. See [`models`](#models).
- `status` - The status of the inference profile. `ACTIVE` means that the inference profile is available to use.
- `type` - The type of the inference profile. `SYSTEM_DEFINED` means that the inference profile is defined by Amazon Bedrock.
- `updated_at` - The time at which the inference profile was last updated.

### `models`

- `model_arn` - The Amazon Resource Name (ARN) of the model.
