---
subcategory: "Bedrock"
layout: "aws"
page_title: "AWS: aws_bedrock_inference_profiles"
description: |-
  Terraform data source for managing AWS Bedrock Inference Profiles.
---

# Data Source: aws_bedrock_inference_profiles

Terraform data source for managing AWS Bedrock AWS Bedrock Inference Profiles.

## Example Usage

### Basic Usage

```terraform
data "aws_bedrock_inference_profiles" "test" {}
```

## Argument Reference

There are no arguments available for this data source.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

- `arns` - List of inference profile summary ARNs.
