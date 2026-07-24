---
subcategory: "SageMaker AI"
layout: "aws"
page_title: "AWS: aws_sagemaker_algorithm"
description: |-
  Lists SageMaker AI Algorithm resources.
---

# List Resource: aws_sagemaker_algorithm

Lists SageMaker AI Algorithm resources.

## Example Usage

```terraform
list "aws_sagemaker_algorithm" "example" {
  provider = aws
}
```

## Argument Reference

This list resource supports the following arguments:

* `region` - (Optional) [Region](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints) to query. Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
