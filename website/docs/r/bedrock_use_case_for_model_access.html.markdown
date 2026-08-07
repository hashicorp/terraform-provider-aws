---
subcategory: "Bedrock"
layout: "aws"
page_title: "AWS: aws_bedrock_use_case_for_model_access"
description: |-
  Manages an AWS Bedrock Use Case For Model Access.
---

# Resource: aws_bedrock_use_case_for_model_access

Manages an AWS Bedrock Use Case For Model Access.

**This is an advanced resource** and has special caveats to be aware of when using it. Please read this document in its entirety before using this resource.

The `aws_bedrock_use_case_for_model_access` resource behaves differently from normal resources in that if an Use Case For Model Access already exists, Terraform does not _create_ this resource, but instead "adopts" it into management if it is the same. As the ability to update doesn't exist, changes compared to the existing resource generate an error.
If no Use Case For Model Access exists, Terraform creates a new Use Case For Model Access.
By default, `terraform destroy` does not delete the Use Case For Model Access but does remove the resource from Terraform state. Real deletion does not exist for this resource.

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

## Attribute Reference

This resource exports no additional attributes.

## Import

In Terraform v1.12.0 and later, the [`import` block](https://developer.hashicorp.com/terraform/language/import) can be used with the `identity` attribute. For example:

```terraform
import {
  to       = aws_bedrock_use_case_for_model_access.example
  identity = {}
}

resource "aws_bedrock_use_case_for_model_access" "example" {}
```

### Identity Schema

#### Optional

* `account_id` (String) AWS Account where this resource is managed.

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import AWS Bedrock Use Case For Model Access resources using the AWS account ID. For example:

```terraform
import {
  to = aws_bedrock_use_case_for_model_access.example
  id = "123456789012"
}
```

Using `terraform import`, import AWS Bedrock Use Case For Model Access resources using the AWS account ID. For example:

```console
% terraform import aws_bedrock_use_case_for_model_access.example 123456789012
```
