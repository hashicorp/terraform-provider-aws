---
subcategory: "SageMaker AI"
layout: "aws"
page_title: "AWS: aws_sagemaker_labeling_job"
description: |-
  Manage an Amazon SageMaker labeling job.
---

# Resource: aws_sagemaker_labeling_job

Manage an Amazon SageMaker labeling job.

## Example Usage

Basic usage:

```terraform
resource "aws_sagemaker_labeling_job" "example" {
  labeling_job_name = "my-labeling-job"
}
```

## Argument Reference

This resource supports the following arguments:

* `labeling_job_name` - (Required) Name of the labeling job
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `tags` - (Optional) A mapping of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

### output_config

* `s3_output_path` - (Required) Amazon S3 output path.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `labeling_job_arn` - The Amazon Resource Name (ARN) of the labeling job.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import labeling jobs using the `labeling_job_name`. For example:

```terraform
import {
  to = aws_sagemaker_labeling_job.example
  id = "my-labeling_job"
}
```

Using `terraform import`, import labeling jobs using the `labeling_job_name`. For example:

```console
% terraform import aws_sagemaker_labeling_job.example my-labeling_job
```