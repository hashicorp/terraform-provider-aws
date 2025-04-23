---
subcategory: "SageMaker AI"
layout: "aws"
page_title: "AWS: aws_sagemaker_model_package_group_policy"
description: |-
  Provides a SageMaker AI Model Package Group Policy resource.
---

# Resource: aws_sagemaker_model_package_group_policy

Provides a SageMaker AI Model Package Group Policy resource.

## Example Usage

### Basic usage

```terraform
data "aws_caller_identity" "current" {}

data "aws_iam_policy_document" "example" {
  statement {
    sid       = "AddPermModelPackageGroup"
    actions   = ["sagemaker:DescribeModelPackage", "sagemaker:ListModelPackages"]
    resources = [aws_sagemaker_model_package_group.example.arn]
    principals {
      identifiers = [data.aws_caller_identity.current.account_id]
      type        = "AWS"
    }
  }
}

resource "aws_sagemaker_model_package_group" "example" {
  model_package_group_name = "example"
}

resource "aws_sagemaker_model_package_group_policy" "example" {
  model_package_group_name = aws_sagemaker_model_package_group.example.model_package_group_name
  resource_policy          = jsonencode(jsondecode(data.aws_iam_policy_document.example.json))
}
```

## Argument Reference

This resource supports the following arguments:

* `model_package_group_name` - (Required) The name of the model package group.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - The name of the Model Package Package Group.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import SageMaker AI Model Package Groups using the `name`. For example:

```terraform
import {
  to = aws_sagemaker_model_package_group_policy.example
  id = "example"
}
```

Using `terraform import`, import SageMaker AI Model Package Groups using the `name`. For example:

```console
% terraform import aws_sagemaker_model_package_group_policy.example example
```
