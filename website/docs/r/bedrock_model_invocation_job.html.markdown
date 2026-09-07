---
subcategory: "Bedrock"
layout: "aws"
page_title: "AWS: aws_bedrock_model_invocation_job"
description: |-
  Manages an Amazon Bedrock model invocation job.
---

# Resource: aws_bedrock_model_invocation_job

Manages an Amazon Bedrock model invocation job. A model invocation job runs a foundation model, or a model accessed through an inference profile, against multiple prompts read from Amazon S3, and writes the results back to Amazon S3.

~> Amazon Bedrock does not support permanently deleting a model invocation job. Destroying this resource stops the job (if it hasn't already reached a terminal state) using the [StopModelInvocationJob](https://docs.aws.amazon.com/bedrock/latest/APIReference/API_StopModelInvocationJob.html) API, then removes it from Terraform state. Set `skip_destroy` to leave the job in its current state instead.

## Example Usage

### Basic Usage

```terraform
resource "aws_bedrock_model_invocation_job" "example" {
  job_name = "example-job"
  model_id = "us.amazon.nova-2-lite-v1:0"
  role_arn = aws_iam_role.example.arn

  input_data_config {
    s3_input_data_config {
      s3_uri = "s3://${aws_s3_bucket.example.id}/input/"
    }
  }

  output_data_config {
    s3_output_data_config {
      s3_uri = "s3://${aws_s3_bucket.example.id}/output/"
    }
  }
}
```

## Argument Reference

The following arguments are required:

* `input_data_config` - (Required) Location of the input data for the batch inference job. See [`input_data_config` Block](#input_data_config-block) below.
* `job_name` - (Required) Name for the batch inference job.
* `model_id` - (Required) Identifier of the foundation model, or inference profile, to use for the batch inference job.
* `output_data_config` - (Required) Location where the results of the batch inference job are stored. See [`output_data_config` Block](#output_data_config-block) below.
* `role_arn` - (Required) ARN of the IAM service role that Amazon Bedrock can assume to carry out and manage the batch inference job. See [Create a service role for batch inference](https://docs.aws.amazon.com/bedrock/latest/userguide/batch-iam-sr.html).

The following arguments are optional:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `skip_destroy` - (Optional) Whether to leave the batch inference job in its current state when destroying the resource, instead of stopping it.
* `tags` - (Optional) Map of tags to assign to the batch inference job. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.
* `timeout_duration_in_hours` - (Optional) Number of hours after which to force the batch inference job to time out.
* `vpc_config` - (Optional) Virtual Private Cloud (VPC) configuration for the data used by the batch inference job. See [`vpc_config` Block](#vpc_config-block) below.

### `input_data_config` Block

The `input_data_config` block supports the following arguments:

* `s3_input_data_config` - (Required) Location of the S3 input data. See [`s3_input_data_config` Block](#s3_input_data_config-block) below.

### `s3_input_data_config` Block

The `s3_input_data_config` block supports the following arguments:

* `s3_bucket_owner` - (Optional) ID of the AWS account that owns the S3 bucket containing the input data.
* `s3_input_format` - (Optional) Format of the input data. Valid values: `JSONL`.
* `s3_uri` - (Required) S3 location of the input data.

### `output_data_config` Block

The `output_data_config` block supports the following arguments:

* `s3_output_data_config` - (Required) Location of the S3 output data. See [`s3_output_data_config` Block](#s3_output_data_config-block) below.

### `s3_output_data_config` Block

The `s3_output_data_config` block supports the following arguments:

* `s3_bucket_owner` - (Optional) ID of the AWS account that owns the S3 bucket containing the output data.
* `s3_encryption_key_id` - (Optional) ARN of the KMS key that encrypts the S3 location of the output data.
* `s3_uri` - (Required) S3 location where the results of the batch inference job are stored.

### `vpc_config` Block

The `vpc_config` block supports the following arguments:

* `security_group_ids` - (Required) IDs of the security groups in the VPC to use.
* `subnet_ids` - (Required) IDs of the subnets in the VPC to use.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `end_time` - Time at which the batch inference job ended.
* `error_record_count` - Number of records that failed to process in the batch inference job.
* `job_arn` - ARN of the batch inference job.
* `job_expiration_time` - Time at which the batch inference job times or timed out.
* `model_invocation_type` - Invocation endpoint used for the batch inference job.
* `processed_record_count` - Number of records that have been processed in the batch inference job.
* `status` - Status of the batch inference job.
* `submit_time` - Time at which the batch inference job was submitted.
* `success_record_count` - Number of records that were successfully processed in the batch inference job.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).
* `total_record_count` - Total number of records in the batch inference job.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `5m`)
* `delete` - (Default `5m`)

## Import

In Terraform v1.12.0 and later, the [`import` block](https://developer.hashicorp.com/terraform/language/import) can be used with the `identity` attribute. For example:

```terraform
import {
  to = aws_bedrock_model_invocation_job.example
  identity = {
    "job_arn" = "arn:aws:bedrock:us-west-2:123456789012:model-invocation-job/abcdefgh1234"
  }
}

resource "aws_bedrock_model_invocation_job" "example" {
  ### Configuration omitted for brevity ###
}
```

### Identity Schema

#### Required

- `job_arn` (String) ARN of the batch inference job.

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Bedrock Model Invocation Job using the `job_arn`. For example:

```terraform
import {
  to = aws_bedrock_model_invocation_job.example
  id = "arn:aws:bedrock:us-west-2:123456789012:model-invocation-job/abcdefgh1234"
}
```

Using `terraform import`, import Bedrock Model Invocation Job using the `job_arn`. For example:

```console
% terraform import aws_bedrock_model_invocation_job.example arn:aws:bedrock:us-west-2:123456789012:model-invocation-job/abcdefgh1234
```
