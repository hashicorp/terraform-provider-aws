---
subcategory: "Bedrock"
layout: "aws"
page_title: "AWS: aws_bedrock_inference_profiles"
description: |-
  Terraform data source for managing AWS Bedrock Inference Profiles.
---

# Data Source: aws_bedrock_inference_profiles

Terraform data source for managing AWS Bedrock Inference Profiles.

## Example Usage

### Basic Usage

```terraform
data "aws_bedrock_inference_profiles" "test" {}
```

### Filter by Type

```terraform
data "aws_bedrock_inference_profiles" "test" {
  type = "APPLICATION"
}
```

## Argument Reference

This data source supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `type` - (Optional) Filters for inference profiles that match the type you specify. Valid values are: `SYSTEM_DEFINED`, `APPLICATION`.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

- `inference_profile_summaries` - List of inference profile summary objects. See [`inference_profile_summaries`](#inference_profile_summaries).

### `inference_profile_summaries`

- `created_at` - Time at which the inference profile was created.
- `description` - Description of the inference profile.
- `inference_profile_arn` - Amazon Resource Name (ARN) of the inference profile.
- `inference_profile_id` - Unique identifier of the inference profile.
- `inference_profile_name` - Name of the inference profile.
- `models` - List of information about each model in the inference profile. See [`models` Block](#models).
- `status` - Status of the inference profile. `ACTIVE` means that the inference profile is available to use.
- `type` - Type of the inference profile. `SYSTEM_DEFINED` means that the inference profile is defined by Amazon Bedrock. `APPLICATION` means the inference profile was created by a user.
- `updated_at` - Time at which the inference profile was last updated.

### `models`

- `model_arn` - Amazon Resource Name (ARN) of the model.
