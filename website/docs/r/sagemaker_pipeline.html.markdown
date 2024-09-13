---
subcategory: "SageMaker"
layout: "aws"
page_title: "AWS: aws_sagemaker_pipeline"
description: |-
  Provides a SageMaker Pipeline resource.
---

# Resource: aws_sagemaker_pipeline

Provides a SageMaker Pipeline resource.

## Example Usage

### Basic usage

```terraform
resource "aws_sagemaker_pipeline" "example" {
  pipeline_name         = "example"
  pipeline_display_name = "example"
  role_arn              = aws_iam_role.example.arn

  pipeline_definition = jsonencode({
    Version = "2020-12-01"
    Steps = [{
      Name = "Test"
      Type = "Fail"
      Arguments = {
        ErrorMessage = "test"
      }
    }]
  })
}
```

## Argument Reference

The following arguments are supported:

* `pipeline_name` - (Required) The name of the pipeline.
* `pipeline_description` - (Optional) A description of the pipeline.
* `pipeline_display_name` - (Required) The display name of the pipeline.
* `pipeline_definition` - (Optional) The [JSON pipeline definition](https://aws-sagemaker-mlops.github.io/sagemaker-model-building-pipeline-definition-JSON-schema/) of the pipeline.
* `pipeline_definition_s3_location` - (Optional) The location of the pipeline definition stored in Amazon S3. If specified, SageMaker will retrieve the pipeline definition from this location. see [Pipeline Definition S3 Location](#pipeline-definition-s3-location) details below.
* `role_arn` - (Required) The ARN of the IAM role the pipeline will execute as.
* `parallelism_configuration` - (Optional) This is the configuration that controls the parallelism of the pipeline. If specified, it applies to all runs of this pipeline by default. see [Parallelism Configuration](#parallelism-configuration) details below.
* `tags` - (Optional) A map of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

### Parallelism Configuration

* `max_parallel_execution_steps` - (Required) The max number of steps that can be executed in parallel.

### Pipeline Definition S3 Location

* `bucket` - (Required) Name of the S3 bucket.
* `object_key` - (Required) The object key (or key name) uniquely identifies the object in an S3 bucket.
* `version_id` - (Optional) Version Id of the pipeline definition file. If not specified, Amazon SageMaker will retrieve the latest version.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - The name of the Pipeline.
* `arn` - The Amazon Resource Name (ARN) assigned by AWS to this Pipeline.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import pipelines using the `pipeline_name`. For example:

```terraform
import {
  to = aws_sagemaker_pipeline.test_pipeline
  id = "pipeline"
}
```

Using `terraform import`, import pipelines using the `pipeline_name`. For example:

```console
% terraform import aws_sagemaker_pipeline.test_pipeline pipeline
```
