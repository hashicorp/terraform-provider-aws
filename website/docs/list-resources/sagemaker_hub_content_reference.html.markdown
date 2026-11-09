---
subcategory: "SageMaker AI"
layout: "aws"
page_title: "AWS: aws_sagemaker_hub_content_reference"
description: |-
  Lists SageMaker AI Hub Content Reference resources.
---

# List Resource: aws_sagemaker_hub_content_reference

Lists SageMaker AI Hub Content Reference resources.

## Example Usage

```terraform
list "aws_sagemaker_hub_content_reference" "example" {
  provider = aws
}
```

## Argument Reference

This list resource supports the following arguments:

* `region` - (Optional) [Region](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints) to query. Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
