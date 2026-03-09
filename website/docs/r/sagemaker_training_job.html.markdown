---
subcategory: "SageMaker AI"
layout: "aws"
page_title: "AWS: aws_sagemaker_training_job"
description: |-
  Manages an AWS SageMaker AI Training Job.
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

# Resource: aws_sagemaker_training_job

Manages an AWS SageMaker AI Training Job.

## Example Usage

### Basic Usage

```terraform
resource "aws_sagemaker_training_job" "example" {
}
```

## Argument Reference

The following arguments are required:

* `example_arg` - (Required) Brief description of the required argument.

The following arguments are optional:

* `optional_arg` - (Optional) Brief description of the optional argument.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the Training Job.
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
  to = aws_sagemaker_training_job.example
  identity = {
    training_job_name = "my-training-job"
  }
}

resource "aws_sagemaker_training_job" "example" {
  ### Configuration omitted for brevity ###
}
```

### Identity Schema

#### Required

* `training_job_name` - (String) Name of the Training Job.

#### Optional

* `account_id` (String) AWS Account where this resource is managed.
* `region` (String) Region where this resource is managed.

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import SageMaker AI Training Job using the `training_job_name`. For example:

```terraform
import {
  to = aws_sagemaker_training_job.example
  id = "my-training-job"
}
```

Using `terraform import`, import SageMaker AI Training Job using the `training_job_name`. For example:

```console
% terraform import aws_sagemaker_training_job.example my-training-job
```
