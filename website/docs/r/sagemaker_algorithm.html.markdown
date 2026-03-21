---
subcategory: "SageMaker AI"
layout: "aws"
page_title: "AWS: aws_sagemaker_algorithm"
description: |-
  Manages an AWS SageMaker AI Algorithm.
---

# Resource: aws_sagemaker_algorithm

Manages an AWS SageMaker AI Algorithm.

## Example Usage

### Basic Usage

```terraform
resource "aws_sagemaker_algorithm" "example" {
}
```

## Argument Reference

The following arguments are required:

* `example_arg` - (Required) Brief description of the required argument.

The following arguments are optional:

* `optional_arg` - (Optional) Brief description of the optional argument.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the Algorithm.
* `example_attribute` - Brief description of the attribute.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `30m`)
* `delete` - (Default `30m`)

## Import

In Terraform v1.12.0 and later, the [`import` block](https://developer.hashicorp.com/terraform/language/import) can be used with the `identity` attribute. For example:

```terraform
import {
  to = aws_sagemaker_algorithm.example
  identity = {
    algorithm_name = "example-algorithm"
  }
}

resource "aws_sagemaker_algorithm" "example" {
  ### Configuration omitted for brevity ###
}
```

### Identity Schema

#### Required

* `algorithm_name` - (String) Name of the Algorithm.

#### Optional

* `account_id` (String) AWS Account where this resource is managed.
* `region` (String) Region where this resource is managed.

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import SageMaker AI Algorithm using the `algorithm_name`. For example:

```terraform
import {
  to = aws_sagemaker_algorithm.example
  id = "example-algorithm"
}
```

Using `terraform import`, import SageMaker AI Algorithm using the `algorithm_name`. For example:

```console
% terraform import aws_sagemaker_algorithm.example example-algorithm
```
