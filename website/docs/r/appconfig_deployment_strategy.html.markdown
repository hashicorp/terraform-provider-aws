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

### AppConfig Deployment Strategy

```hcl
resource "aws_appconfig_deployment_strategy" "test" {
  name                           = "test"
  description                    = "test"
  deployment_duration_in_minutes = 3
  final_bake_time_in_minutes     = 4
  growth_factor                  = 10
  growth_type                    = "LINEAR"
  replicate_to                   = "NONE"
  tags = {
    Env = "Test"
  }
}
```

## Argument Reference

The following arguments are supported:

- `name` - (Required, Forces new resource) A name for the deployment strategy. Must be between 1 and 64 characters in length.
- `description` - (Optional) A description of the deployment strategy. Can be at most 1024 characters.
- `deployment_duration_in_minutes` - (Required) Total amount of time for a deployment to last.
- `final_bake_time_in_minutes` - (Optional) The amount of time AWS AppConfig monitors for alarms before considering the deployment to be complete and no longer eligible for automatic roll back.
- `growth_factor` - (Required) The percentage of targets to receive a deployed configuration during each interval.
- `growth_type` - (Optional) The algorithm used to define how percentage grows over time.
- `replicate_to` - (Required) Save the deployment strategy to a Systems Manager (SSM) document.
- `tags` - (Optional) A map of tags to assign to the resource.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

- `arn` - The Amazon Resource Name (ARN) of the AppConfig Deployment Strategy.
- `id` - The AppConfig Deployment Strategy ID

## Import

`aws_appconfig_deployment_strategy` can be imported by the Deployment Strategy ID, e.g.

```
$ terraform import aws_appconfig_deployment_strategy.test 11xxxxx
```
