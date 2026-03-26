---
subcategory: "SageMaker AI"
layout: "aws"
page_title: "AWS: aws_sagemaker_hyper_parameter_tuning_job"
description: |-
  Manages an AWS SageMaker AI Hyper Parameter Tuning Job.
---
<!---
Documentation guidelines:
- Begin resource descriptions with "Manages..."
- Use simple language and avoid jargon
- Focus on brevity and clarity
- Use present tense and active voice
- Don't begin argument/attribute descriptions with "An", "The", "Defines", "Indicates", or "Specifies"
- Boolean arguments should begin with "Whether to"
- Use "example" instead of "test" in examples
--->

# Resource: aws_sagemaker_hyper_parameter_tuning_job

Manages an AWS SageMaker AI Hyper Parameter Tuning Job.

## Example Usage

### Basic Usage

```terraform
resource "aws_sagemaker_hyper_parameter_tuning_job" "example" {
}
```

## Argument Reference

The following arguments are required:

* `example_arg` - (Required) Brief description of the required argument.

The following arguments are optional:

* `optional_arg` - (Optional) Brief description of the optional argument.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the Hyper Parameter Tuning Job.
* `example_attribute` - Brief description of the attribute.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `60m`)
* `update` - (Default `180m`)
* `delete` - (Default `90m`)

## Import

In Terraform v1.12.0 and later, the [`import` block](https://developer.hashicorp.com/terraform/language/import) can be used with the `identity` attribute. For example:

```terraform
import {
  to = aws_sagemaker_hyper_parameter_tuning_job.example
  identity = {
    hyper_parameter_tuning_job_name = "example-hyper-parameter-tuning-job"
  }
}

resource "aws_sagemaker_hyper_parameter_tuning_job" "example" {
  ### Configuration omitted for brevity ###
}
```

### Identity Schema

#### Required
* `hyper_parameter_tuning_job_name` (String) Name of the Hyper Parameter Tuning Job.

#### Optional
* `account_id` (String) AWS Account where this resource is managed.
* `region` (String) Region where this resource is managed.
