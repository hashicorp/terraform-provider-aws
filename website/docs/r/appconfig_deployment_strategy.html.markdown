---
subcategory: "AppConfig"
layout: "aws"
page_title: "AWS: aws_appconfig_deployment_strategy"
description: |-
  Provides an AppConfig Deployment Strategy resource.
---

# Resource: aws_appconfig_deployment_strategy

Provides an AppConfig Deployment Strategy resource.

## Example Usage

```terraform
resource "aws_appconfig_deployment_strategy" "example" {
  name                           = "example-deployment-strategy-tf"
  description                    = "Example Deployment Strategy"
  deployment_duration_in_minutes = 3
  final_bake_time_in_minutes     = 4
  growth_factor                  = 10
  growth_type                    = "LINEAR"
  replicate_to                   = "NONE"

  tags = {
    Type = "AppConfig Deployment Strategy"
  }
}
```

## Argument Reference

This resource supports the following arguments:

* `deployment_duration_in_minutes` - (Required) Total amount of time for a deployment to last. Minimum value of 0, maximum value of 1440.
* `growth_factor` - (Required) Percentage of targets to receive a deployed configuration during each interval. Minimum value of 1.0, maximum value of 100.0.
* `name` - (Required, Forces new resource) Name for the deployment strategy. Must be between 1 and 64 characters in length.
* `replicate_to` - (Required, Forces new resource) Where to save the deployment strategy. Valid values: `NONE` and `SSM_DOCUMENT`.
* `description` - (Optional) Description of the deployment strategy. Can be at most 1024 characters.
* `final_bake_time_in_minutes` - (Optional) Amount of time AWS AppConfig monitors for alarms before considering the deployment to be complete and no longer eligible for automatic roll back. Minimum value of 0, maximum value of 1440.
* `growth_type` - (Optional) Algorithm used to define how percentage grows over time. Valid value: `LINEAR` and `EXPONENTIAL`. Defaults to `LINEAR`.
* `tags` - (Optional) Map of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - AppConfig deployment strategy ID.
* `arn` - ARN of the AppConfig Deployment Strategy.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import AppConfig Deployment Strategies using their deployment strategy ID. For example:

```terraform
import {
  to = aws_appconfig_deployment_strategy.example
  id = "11xxxxx"
}
```

Using `terraform import`, import AppConfig Deployment Strategies using their deployment strategy ID. For example:

```console
% terraform import aws_appconfig_deployment_strategy.example 11xxxxx
```
