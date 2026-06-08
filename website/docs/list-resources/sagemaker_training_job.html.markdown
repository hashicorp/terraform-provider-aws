---
subcategory: "SageMaker AI"
layout: "aws"
page_title: "AWS: aws_sagemaker_training_job"
description: |-
  Lists SageMaker AI Training Job resources.
---

# List Resource: aws_sagemaker_training_job

Lists SageMaker AI Training Job resources.

## Example Usage

```terraform
list "aws_sagemaker_training_job" "example" {
  provider = aws
}
```

## Argument Reference

This list resource supports the following arguments:

* `region` - (Optional) Region to query. Defaults to provider region.
