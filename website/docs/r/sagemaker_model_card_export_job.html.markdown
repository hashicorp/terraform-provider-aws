---
subcategory: "SageMaker AI"
layout: "aws"
page_title: "AWS: aws_sagemaker_model_card_export_job"
description: |-
  Manage an Amazon SageMaker Model Card export job.
---

# Resource: aws_sagemaker_model_card_export_job

Manage an Amazon SageMaker Model Card export job.

## Example Usage

Basic usage:

```terraform
resource "aws_sagemaker_model_card_export_job" "example" {
  model_card_export_job_name = "my-model-card-export-job"
  model_card_name            = aws_sagemaker_model_card.example.model_card_name

  output_config {
    s3_output_path = "s3://${aws_s3_bucket.test.example}/"
  }
}
```

## Argument Reference

This resource supports the following arguments:

* `model_card_name` - (Required) Name of the model card.
* `model_card_export_job_name` - (Required) Name of the model card export job.
* `model_card_export_job_version` - (Optional) Version of the model card that the model card export job exports.
* `output_config` - (Required) Export output details. Fields are documented below.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference)

### output_config

* `s3_output_path` - (Required) Amazon S3 output path.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `export_artifacts` - Exported model card artifacts.
    * `s3_export_artifacts` - Amazon S3 URI of the exported model artifacts.
* `model_card_export_job_arn` - The Amazon Resource Name (ARN) of the model card export job.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

- `create` - (Default `15m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import model card export jobs using the `model_card_export_job_arn`. For example:

```terraform
import {
  to = aws_sagemaker_model_card_export_job.example
  id = "arn:aws:sagemaker:us-west-2:123456789012:model-card/my-model-card/export-job/my-model-card-export-job"
}
```

Using `terraform import`, import model card export jobs using the `model_card_export_job_arn`. For example:

```console
% terraform import aws_sagemaker_model_card_export_job.example arn:aws:sagemaker:us-west-2:123456789012:model-card/my-model-card/export-job/my-model-card-export-job
```