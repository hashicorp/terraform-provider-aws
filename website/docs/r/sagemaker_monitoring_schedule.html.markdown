---
subcategory: "SageMaker"
layout: "aws"
page_title: "AWS: aws_sagemaker_monitoring_schedule"
description: |-
  Provides a SageMaker Monitoring Schedule resource.
---

# Resource: aws_sagemaker_monitoring_schedule

Provides a SageMaker monitoring schedule resource.

## Example Usage

Basic usage:

```terraform
resource "aws_sagemaker_monitoring_schedule" "test" {
  name = "my-monitoring-schedule"

  monitoring_schedule_config {
    monitoring_job_definition_name = aws_sagemaker_data_quality_job_definition.test.name
    monitoring_type                = "DataQuality"
  }
}
```

## Argument Reference

This resource supports the following arguments:

* `monitoring_schedule_config` - (Required) The configuration object that specifies the monitoring schedule and defines the monitoring job. Fields are documented below.
* `name` - (Optional) The name of the monitoring schedule. The name must be unique within an AWS Region within an AWS account. If omitted, Terraform will assign a random, unique name.
* `tags` - (Optional) A mapping of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

### monitoring_schedule_config

* `monitoring_job_definition_name` - (Required) The name of the monitoring job definition to schedule.
* `monitoring_type` - (Required) The type of the monitoring job definition to schedule. Valid values are `DataQuality`, `ModelQuality`, `ModelBias` or `ModelExplainability`
* `schedule_config` - (Optional) Configures the monitoring schedule. Fields are documented below.

#### schedule_config

* `schedule_expression` - (Required) A cron expression that describes details about the monitoring schedule. For example, and hourly schedule would be `cron(0 * ? * * *)`.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - The Amazon Resource Name (ARN) assigned by AWS to this monitoring schedule.
* `name` - The name of the monitoring schedule.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import monitoring schedules using the `name`. For example:

```terraform
import {
  to = aws_sagemaker_monitoring_schedule.test_monitoring_schedule
  id = "monitoring-schedule-foo"
}
```

Using `terraform import`, import monitoring schedules using the `name`. For example:

```console
% terraform import aws_sagemaker_monitoring_schedule.test_monitoring_schedule monitoring-schedule-foo
```
