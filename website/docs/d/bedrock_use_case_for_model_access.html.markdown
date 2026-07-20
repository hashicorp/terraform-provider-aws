---
subcategory: "Bedrock"
layout: "aws"
page_title: "AWS: aws_bedrock_use_case_for_model_access"
description: |-
  Provides details about an AWS Bedrock Use Case For Model Access.
---

# Data Source: aws_bedrock_use_case_for_model_access

Provides details about an AWS Bedrock Use Case For Model Access.

## Example Usage

### Basic Usage

```terraform
data "aws_bedrock_use_case_for_model_access" "example" {
}
```

## Argument Reference

The following arguments are required:

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `form_data` - Form data as JSON from the Anthropic first time user request.
