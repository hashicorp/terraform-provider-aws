---
subcategory: "App Runner"
layout: "aws"
page_title: "AWS: aws_apprunner_default_auto_scaling_configuration_version"
description: |-
  Manages the default App Runner auto scaling configuration.
---

# Resource: aws_apprunner_default_auto_scaling_configuration_version

Manages the default App Runner auto scaling configuration.
When creating or updating this resource the existing default auto scaling configuration will be set to non-default automatically.
When creating or updating this resource the configuration is automatically assigned as the default to the new services you create in the future. The new default designation doesn't affect the associations that were previously set for existing services.
Each account can have only one default auto scaling configuration per Region.

## Example Usage

```terraform
resource "aws_apprunner_auto_scaling_configuration_version" "example" {
  auto_scaling_configuration_name = "example"

  max_concurrency = 50
  max_size        = 10
  min_size        = 2
}

resource "aws_apprunner_default_auto_scaling_configuration_version" "example" {
  auto_scaling_configuration_arn = aws_apprunner_auto_scaling_configuration_version.example.arn
}
```

## Argument Reference

The following arguments supported:

* `auto_scaling_configuration_arn` - (Required) The ARN of the App Runner auto scaling configuration that you want to set as the default.

## Attribute Reference

This resource exports no additional attributes.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import App Runner default auto scaling configurations using the current Region. For example:

```terraform
import {
  to = aws_apprunner_default_auto_scaling_configuration_version.example
  id = "us-west-2"
}
```

Using `terraform import`, import App Runner default auto scaling configurations using the current Region. For example:

```console
% terraform import aws_apprunner_default_auto_scaling_configuration_version.example us-west-2
```
