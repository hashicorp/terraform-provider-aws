---
subcategory: "SageMaker AI"
layout: "aws"
page_title: "AWS: aws_sagemaker_hyper_parameter_tuning_job"
description: |-
  Lists SageMaker Hyper Parameter Tuning Job resources.
---

# List Resource: aws_sagemaker_hyper_parameter_tuning_job

Lists SageMaker Hyper Parameter Tuning Job resources.

## Example Usage

```terraform
list "aws_sagemaker_hyper_parameter_tuning_job" "example" {
  provider = aws
}
```

## Argument Reference

This list resource supports the following arguments:

* `region` - (Optional) Region to query. Defaults to provider region.
