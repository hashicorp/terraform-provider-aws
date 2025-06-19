---
subcategory: "Bedrock"
layout: "aws"
page_title: "AWS: aws_bedrock_inference_profile"
description: |-
  Terraform resource for managing an AWS Bedrock Inference Profile.
---

# Resource: aws_bedrock_inference_profile

Terraform resource for managing an AWS Bedrock Inference Profile.

## Example Usage

### Basic Usage

```terraform
data "aws_caller_identity" "current" {}

resource "aws_bedrock_inference_profile" "example" {
  name        = "Claude Sonnet for Project 123"
  description = "Profile with tag for cost allocation tracking"

  model_source {
    copy_from = "arn:aws:bedrock:us-west-2::foundation-model/anthropic.claude-3-5-sonnet-20241022-v2:0"

    # Include account ID to use inference profiles
    # copy_from = "arn:aws:bedrock:eu-central-1:${data.aws_caller_identity.current.account_id}:inference-profile/eu.anthropic.claude-3-5-sonnet-20240620-v1:0"
  }

  tags = {
    ProjectID = "123"
  }
}
```

## Argument Reference

The following arguments are required:

* `name` - (Required) The name of the inference profile.
* `model_source` - (Required) The source of the model this inference profile will track metrics and cost for. See [`model_source`](#model_source).

The following arguments are optional:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `description` - (Optional) The description of the inference profile.
* `tags` - (Optional) Key-value mapping of resource tags for the inference profile.

### `model_source`

- `copy_from` - The Amazon Resource Name (ARN) of the model.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

- `arn` - The Amazon Resource Name (ARN) of the inference profile.
- `id` - The unique identifier of the inference profile.
- `name` - The unique identifier of the inference profile.
- `models` - A list of information about each model in the inference profile. See [`models`](#models).
- `status` - The status of the inference profile. `ACTIVE` means that the inference profile is available to use.
- `type` - The type of the inference profile. `SYSTEM_DEFINED` means that the inference profile is defined by Amazon Bedrock. `APPLICATION` means that the inference profile is defined by the user.
- `created_at` - The time at which the inference profile was created.
- `description` - The description of the inference profile.
- `updated_at` - The time at which the inference profile was last updated.

### `models`

- `model_arn` - The Amazon Resource Name (ARN) of the model.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `5m`)
* `update` - (Default `5m`)
* `delete` - (Default `5m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Bedrock Inference Profile using the `example_id_arg`. For example:

```terraform
import {
  to = aws_bedrock_inference_profile.example
  id = "inference_profile-id-12345678"
}
```

Using `terraform import`, import Bedrock Inference Profile using the `example_id_arg`. For example:

```console
% terraform import aws_bedrock_inference_profile.example inference_profile-id-12345678
```
