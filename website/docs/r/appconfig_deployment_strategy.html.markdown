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

The following arguments are supported:

* `deployment_duration_in_minutes` - (Required) Total amount of time for a deployment to last. Minimum value of 0, maximum value of 1440.
* `growth_factor` - (Required) The percentage of targets to receive a deployed configuration during each interval. Minimum value of 1.0, maximum value of 100.0.
* `name` - (Required, Forces new resource) A name for the deployment strategy. Must be between 1 and 64 characters in length.
* `replicate_to` - (Required, Forces new resource) Where to save the deployment strategy. Valid values: `NONE` and `SSM_DOCUMENT`.
* `description` - (Optional) A description of the deployment strategy. Can be at most 1024 characters.
* `final_bake_time_in_minutes` - (Optional) The amount of time AWS AppConfig monitors for alarms before considering the deployment to be complete and no longer eligible for automatic roll back. Minimum value of 0, maximum value of 1440.
* `growth_type` - (Optional) The algorithm used to define how percentage grows over time. Valid value: `LINEAR` and `EXPONENTIAL`. Defaults to `LINEAR`.
* `tags` - (Optional) A map of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The AppConfig deployment strategy ID.
* `arn` - The Amazon Resource Name (ARN) of the AppConfig Deployment Strategy.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block).

## Import

AppConfig Deployment Strategies can be imported by using their deployment strategy ID, e.g.,

```
$ terraform import aws_appconfig_deployment_strategy.example 11xxxxx
```
