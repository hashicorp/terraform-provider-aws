---
subcategory: "Glue"
layout: "aws"
page_title: "AWS: aws_glue_trigger"
description: |-
  Manages a Glue Trigger resource.
---

# Resource: aws_glue_trigger

Manages a Glue Trigger resource.

## Example Usage

### Conditional Trigger

```terraform
resource "aws_glue_trigger" "example" {
  name = "example"
  type = "CONDITIONAL"

  actions {
    job_name = aws_glue_job.example1.name
  }

  predicate {
    conditions {
      job_name = aws_glue_job.example2.name
      state    = "SUCCEEDED"
    }
  }
}
```

### On-Demand Trigger

```terraform
resource "aws_glue_trigger" "example" {
  name = "example"
  type = "ON_DEMAND"

  actions {
    job_name = aws_glue_job.example.name
  }
}
```

### Scheduled Trigger

```terraform
resource "aws_glue_trigger" "example" {
  name     = "example"
  schedule = "cron(15 12 * * ? *)"
  type     = "SCHEDULED"

  actions {
    job_name = aws_glue_job.example.name
  }
}
```

### Conditional Trigger with Crawler Action

**Note:** Triggers can have both a crawler action and a crawler condition, just no example provided.

```terraform
resource "aws_glue_trigger" "example" {
  name = "example"
  type = "CONDITIONAL"

  actions {
    crawler_name = aws_glue_crawler.example1.name
  }

  predicate {
    conditions {
      job_name = aws_glue_job.example2.name
      state    = "SUCCEEDED"
    }
  }
}
```

### Conditional Trigger with Crawler Condition

**Note:** Triggers can have both a crawler action and a crawler condition, just no example provided.

```terraform
resource "aws_glue_trigger" "example" {
  name = "example"
  type = "CONDITIONAL"

  actions {
    job_name = aws_glue_job.example1.name
  }

  predicate {
    conditions {
      crawler_name = aws_glue_crawler.example2.name
      crawl_state  = "SUCCEEDED"
    }
  }
}
```

## Argument Reference

This resource supports the following arguments:

* `actions` - (Required) List of actions initiated by this trigger when it fires. See [`actions`](#actions-block) below.
* `description` - (Optional) Description of the new trigger.
* `enabled` - (Optional) Start the trigger. Defaults to `true`.
* `event_batching_condition` - (Optional) Batch condition that must be met (specified number of events received or batch time window expired) before EventBridge event trigger fires. See [`event_batching_condition`](#event_batching_condition-block).
* `name` - (Required) Name of the trigger.
* `predicate` - (Optional) Predicate to specify when the new trigger should fire. Required when trigger type is `CONDITIONAL`. See [`predicate`](#predicate-block) below.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `schedule` - (Optional) Cron expression used to specify the schedule. [Time-Based Schedules for Jobs and Crawlers](https://docs.aws.amazon.com/glue/latest/dg/monitor-data-warehouse-schedule.html)
* `start_on_creation` - (Optional) Set to true to start `SCHEDULED` and `CONDITIONAL` triggers when created. True is not supported for `ON_DEMAND` triggers.
* `tags` - (Optional) Key-value map of resource tags. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.
* `type` - (Required) Type of trigger. Valid values are `CONDITIONAL`, `EVENT`, `ON_DEMAND`, and `SCHEDULED`.
* `workflow_name` - (Optional) Workflow to which the trigger should be associated to. Every workflow graph (DAG) needs a starting trigger (`ON_DEMAND` or `SCHEDULED` type) and can contain multiple additional `CONDITIONAL` triggers.

### `actions` Block

* `arguments` - (Optional) Arguments to be passed to the job. You can specify arguments here that your own job-execution script consumes, as well as arguments that AWS Glue itself consumes.
* `crawler_name` - (Optional) Name of the crawler to be executed. Conflicts with `job_name`.
* `job_name` - (Optional) Name of a job to be executed. Conflicts with `crawler_name`.
* `notification_property` - (Optional) Configuration properties of a job run notification. See [`notification_property`](#notification_property-block) details below.
* `security_configuration` - (Optional) Name of the Security Configuration structure to be used with this action.
* `timeout` - (Optional) Job run timeout in minutes. It overrides the timeout value of the job.

#### `notification_property` Block

* `notify_delay_after` - (Optional) After a job run starts, the number of minutes to wait before sending a job run delay notification.

### `predicate` Block

* `conditions` - (Required) List of the conditions that determine when the trigger will fire. See [`conditions`](#conditions-block).
* `logical` - (Optional) How to handle multiple conditions. Defaults to `AND`. Valid values are `AND` or `ANY`.

#### `conditions` Block

* `crawl_state` - (Optional) Condition crawl state. Currently, the values supported are `RUNNING`, `SUCCEEDED`, `CANCELLED`, and `FAILED`. If this is specified, `crawler_name` must also be specified. Conflicts with `state`.
* `crawler_name` - (Optional) Name of the crawler to watch. If this is specified, `crawl_state` must also be specified. Conflicts with `job_name`.
* `job_name` - (Optional) Name of the job to watch. If this is specified, `state` must also be specified. Conflicts with `crawler_name`.
* `logical_operator` - (Optional) Logical operator. Defaults to `EQUALS`.
* `state` - (Optional) Condition job state. Currently, the values supported are `SUCCEEDED`, `STOPPED`, `TIMEOUT` and `FAILED`. If this is specified, `job_name` must also be specified. Conflicts with `crawler_state`.

### `event_batching_condition` Block

* `batch_size` - (Required) Number of events that must be received from Amazon EventBridge before EventBridge event trigger fires.
* `batch_window` - (Optional) Window of time in seconds after which EventBridge event trigger fires. Window starts when first event is received. Default value is `900`.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of Glue Trigger
* `id` - Trigger name
* `state` - Current state of the trigger.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

- `create` - (Default `5m`)
- `update` - (Default `5m`)
- `delete` - (Default `5m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Glue Triggers using `name`. For example:

```terraform
import {
  to = aws_glue_trigger.MyTrigger
  id = "MyTrigger"
}
```

Using `terraform import`, import Glue Triggers using `name`. For example:

```console
% terraform import aws_glue_trigger.MyTrigger MyTrigger
```
