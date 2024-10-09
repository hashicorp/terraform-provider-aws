---
subcategory: "Rekognition"
layout: "aws"
page_title: "AWS: aws_rekognition_project"
description: |-
  Terraform resource for managing an AWS Rekognition Project.
---

# Resource: aws_rekognition_project

Terraform resource for managing an AWS Rekognition Project.

## Example Usage

```terraform
resource "aws_rekognition_project" "example" {
  name        = "example-project"
  auto_update = "ENABLED"
  feature     = "CONTENT_MODERATION"
}
```

## Argument Reference

The following arguments are required:

* `name` - (Required) Desired name of the project

The following arguments are optional:

* `auto_update` - (Optional) Specify if automatic retraining should occur. Valid values are `ENABLED` or `DISABLED`. Defaults to `DISABLED`
* `feature` - (Optional) Specify the feature being customized. Valid values are `CONTENT_MODERATION` or `CUSTOM_LABELS`. Defaults to `CUSTOM_LABELS`

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the Project.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `10m`)
* `delete` - (Default `10m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Rekognition Project using the `example_id_arg`. For example:

```terraform
import {
  to = aws_rekognition_project.example
  id = "project-id-12345678"
}
```

Using `terraform import`, import Rekognition Project using the `name`. For example:

```console
% terraform import aws_rekognition_project.example project-id-12345678
```
