---
subcategory: "Bedrock"
layout: "aws"
page_title: "AWS: aws_bedrock_use_case_for_model_access"
description: |-
  Manages an AWS Bedrock Use Case For Model Access.
---

# Resource: aws_bedrock_use_case_for_model_access

Manages an AWS Bedrock Use Case For Model Access.

## Example Usage

### Basic Usage

```terraform
resource "aws_bedrock_use_case_for_model_access" "example" {
  form_data = jsonencode({
    "companyName"         = "AWS Provider",
    "companyWebsite"      = "https://www.test.com",
    "intendedUsers"       = "0",
    "industryOption"      = "Energy",
    "otherIndustryOption" = "",
    "useCases"            = ". - Generating developer documentation\n- Code generation/refactoring\n- Summarization of issues / documents"
  })
}
```

## Argument Reference

The following arguments are required:

* `form_data` - (Required) Brief description of the required argument.

The following arguments are optional:

* `region` - (Optional) Brief description of the optional argument.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `30m`)
* `update` - (Default `30m`)
* `delete` - (Default `30m`)
