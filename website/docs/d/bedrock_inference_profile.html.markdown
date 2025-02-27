---
subcategory: "Bedrock"
layout: "aws"
page_title: "AWS: aws_bedrock_inference_profile"
description: |-
  Terraform data source for managing an AWS Bedrock Inference Profile.
---

# Data Source: aws_bedrock_inference_profile

Terraform data source for managing an AWS Bedrock Inference Profile.

## Example Usage

### Basic Usage

```terraform
data "aws_bedrock_inference_profiles" "test" {}

data "aws_bedrock_inference_profile" "test" {
  inference_profile_id = data.aws_bedrock_inference_profiles.test.inference_profile_summaries[0].inference_profile_id
}
```

## Argument Reference

The following argument are required:

- `inference_profile_id` – (Required) Inference Profile identifier.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

- `inference_profile_arn` - The Amazon Resource Name (ARN) of the inference profile.
- `inference_profile_name` - The unique identifier of the inference profile.
- `models` - A list of information about each model in the inference profile. See [`models`](#models).
- `status` - The status of the inference profile. `ACTIVE` means that the inference profile is available to use.
- `type` - The type of the inference profile. `SYSTEM_DEFINED` means that the inference profile is defined by Amazon Bedrock. `APPLICATION` means that the inference profile is defined by the user.
- `created_at` - The time at which the inference profile was created.
- `description` - The description of the inference profile.
- `updated_at` - The time at which the inference profile was last updated.

### `models`

- `model_arn` - The Amazon Resource Name (ARN) of the model.
