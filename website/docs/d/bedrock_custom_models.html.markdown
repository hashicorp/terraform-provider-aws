---
subcategory: "Bedrock"
layout: "aws"
page_title: "AWS: aws_bedrock_custom_models"
description: |-
  Returns a list of Amazon Bedrock custom models.
---

# Data Source: aws_bedrock_custom_models

Returns a list of Amazon Bedrock custom models.

## Example Usage

```terraform
data "aws_bedrock_custom_models" "test" {}
```

## Argument Reference

None.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `model_summaries` - Model summaries.
    * `creation_time` - Creation time of the model.
    * `model_arn` - The ARN of the custom model.
    * `model_name` - The name of the custom model.
