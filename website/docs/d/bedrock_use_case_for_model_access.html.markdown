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

This data source supports the following arguments:

* `account` - (Optional) AWS account ID where this data source will be read. Defaults to the account of the provider's credentials. See the [Per-Resource Account guide](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/guides/per-resource-account).

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `form_data` - Form data as JSON from the Anthropic first time user request.
