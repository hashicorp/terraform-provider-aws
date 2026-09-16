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

This data source supports the following arguments:

* `inference_profile_id` - (Required) Inference Profile identifier.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

- `created_at` - Time at which the inference profile was created.
- `description` - Description of the inference profile.
- `inference_profile_arn` - ARN of the inference profile.
- `inference_profile_name` - Unique identifier of the inference profile.
- `models` - List of information about each model in the inference profile. See [`models`](#models).
- `status` - Status of the inference profile. `ACTIVE` means that the inference profile is available to use.
- `type` - Type of the inference profile. `SYSTEM_DEFINED` means that the inference profile is defined by Amazon Bedrock. `APPLICATION` means that the inference profile is defined by the user.
- `updated_at` - Time at which the inference profile was last updated.

### `models` Block

- `model_arn` - ARN of the model.
