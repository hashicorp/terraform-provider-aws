---
subcategory: "SageMaker"
layout: "aws"
page_title: "AWS: aws_sagemaker_model_package_group_policy"
description: |-
  Provides a SageMaker Model Package Group Policy resource.
---

# Resource: aws_sagemaker_model_package_group_policy

Provides a SageMaker Model Package Group Policy resource.

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

The following arguments are supported:

* `model_package_group_name` - (Required) The name of the model package group.


## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The name of the Model Package Package Group.

## Import

SageMaker Code Model Package Groups can be imported using the `name`, e.g.,

```
$ terraform import aws_sagemaker_model_package_group_policy.example example
```
