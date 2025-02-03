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

* `name` - (Required) Desired name of the project.

The following arguments are optional:

* `auto_update` - (Optional) Specify if automatic retraining should occur. Valid values are `ENABLED` or `DISABLED`. Defaults to `DISABLED`.
* `feature` - (Optional) Specify the feature being customized. Valid values are `CONTENT_MODERATION` or `CUSTOM_LABELS`. Defaults to `CUSTOM_LABELS`.
* `tags` - (Optional) Map of tags assigned to the resource. If configured with a provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the Project.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

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
