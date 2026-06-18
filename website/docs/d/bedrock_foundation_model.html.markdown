---
subcategory: "Bedrock"
layout: "aws"
page_title: "AWS: aws_bedrock_foundation_model"
description: |-
  Terraform data source for managing an AWS Bedrock Foundation Model.
---

# Data Source: aws_bedrock_foundation_model

Terraform data source for managing an AWS Bedrock Foundation Model.

## Example Usage

### Basic Usage

```terraform
data "aws_bedrock_foundation_models" "test" {}

data "aws_bedrock_foundation_model" "test" {
  model_id = data.aws_bedrock_foundation_models.test.model_summaries[0].model_id
}
```

## Argument Reference

This data source supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `model_id` - (Required) Model identifier.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `customizations_supported` - Customizations that the model supports.
* `inference_types_supported` - Inference types that the model supports.
* `input_modalities` - Input modalities that the model supports.
* `model_arn` - Model ARN.
* `model_name` - Model name.
* `output_modalities` - Output modalities that the model supports.
* `provider_name` - Model provider name.
* `response_streaming_supported` - Indicates whether the model supports streaming.
