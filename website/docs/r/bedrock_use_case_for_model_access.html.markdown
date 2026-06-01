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

* `form_data` - (Required) Form data from the Anthropic first time user request. See also the example [payload](https://docs.aws.amazon.com/bedrock/latest/APIReference/API_PutUseCaseForModelAccess.html#API_PutUseCaseForModelAccess_Examples).

The following arguments are optional:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `30m`)
* `update` - (Default `30m`)
* `delete` - (Default `30m`)
