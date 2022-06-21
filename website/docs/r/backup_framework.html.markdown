---
subcategory: "Backup"
layout: "aws"
page_title: "AWS: aws_backup_framework"
description: |-
  Provides an AWS Backup Framework resource.
---

# Resource: aws_backup_framework

Provides an AWS Backup Framework resource.

~> **Note:** For the Deployment Status of the Framework to be successful, please turn on resource tracking to enable AWS Config recording to track configuration changes of your backup resources. This can be done from the AWS Console.

## Example Usage

```terraform
resource "aws_backup_framework" "Example" {
  name        = "exampleFramework"
  description = "this is an example framework"

  control {
    name = "BACKUP_RECOVERY_POINT_MINIMUM_RETENTION_CHECK"

    input_parameter {
      name  = "requiredRetentionDays"
      value = "35"
    }
  }

  control {
    name = "BACKUP_PLAN_MIN_FREQUENCY_AND_MIN_RETENTION_CHECK"

    input_parameter {
      name  = "requiredFrequencyUnit"
      value = "hours"
    }

    input_parameter {
      name  = "requiredRetentionDays"
      value = "35"
    }

    input_parameter {
      name  = "requiredFrequencyValue"
      value = "1"
    }
  }

  control {
    name = "BACKUP_RECOVERY_POINT_ENCRYPTED"
  }

  control {
    name = "BACKUP_RESOURCES_PROTECTED_BY_BACKUP_PLAN"

    scope {
      compliance_resource_types = [
        "EBS"
      ]
    }
  }

  control {
    name = "BACKUP_RECOVERY_POINT_MANUAL_DELETION_DISABLED"
  }

  tags = {
    "Name" = "Example Framework"
  }
}
```

## Argument Reference

The following arguments are supported:

* `control` - (Required) One or more control blocks that make up the framework. Each control in the list has a name, input parameters, and scope. Detailed below.
* `description` - (Optional) The description of the framework with a maximum of 1,024 characters
* `name` - (Required) The unique name of the framework. The name must be between 1 and 256 characters, starting with a letter, and consisting of letters, numbers, and underscores.
* `tags` - (Optional) Metadata that you can assign to help organize the frameworks you create. If configured with a provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

### Control Arguments
For **control** the following attributes are supported:

* `input_parameter` - (Optional) One or more input parameter blocks. An example of a control with two parameters is: "backup plan frequency is at least daily and the retention period is at least 1 year". The first parameter is daily. The second parameter is 1 year. Detailed below.
* `name` - (Required) The name of a control. This name is between 1 and 256 characters.
* `scope` - (Optional) The scope of a control. The control scope defines what the control will evaluate. Three examples of control scopes are: a specific backup plan, all backup plans with a specific tag, or all backup plans. Detailed below.

### Input Parameter Arguments
For **input_parameter** the following attributes are supported:

* `name` - (Optional) The name of a parameter, for example, BackupPlanFrequency.
* `value` - (Optional) The value of parameter, for example, hourly.

### Scope Arguments
For **scope** the following attributes are supported:

* `compliance_resource_ids` - (Optional) The ID of the only AWS resource that you want your control scope to contain. Minimum number of 1 item. Maximum number of 100 items.
* `compliance_resource_types` - (Optional) Describes whether the control scope includes one or more types of resources, such as EFS or RDS.
* `tags` - (Optional) The tag key-value pair applied to those AWS resources that you want to trigger an evaluation for a rule. A maximum of one key-value pair can be provided.

### Timeouts

The `timeouts` block allows you to specify [timeouts](https://www.terraform.io/docs/configuration/blocks/resources/syntax.html#operation-timeouts) for certain actions:

* `create` - (Defaults to 2 mins) Used when creating the framework
* `update` - (Defaults to 2 mins) Used when updating the framework
* `delete` - (Defaults to 2 mins) Used when deleting the framework

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - The ARN of the backup framework.
* `creation_time` - The date and time that a framework is created, in Unix format and Coordinated Universal Time (UTC).
* `deployment_status` - The deployment status of a framework. The statuses are: `CREATE_IN_PROGRESS` | `UPDATE_IN_PROGRESS` | `DELETE_IN_PROGRESS` | `COMPLETED` | `FAILED`.
* `id` - The id of the backup framework.
* `status` - A framework consists of one or more controls. Each control governs a resource, such as backup plans, backup selections, backup vaults, or recovery points. You can also turn AWS Config recording on or off for each resource. For more information refer to the [AWS documentation for Framework Status](https://docs.aws.amazon.com/aws-backup/latest/devguide/API_DescribeFramework.html#Backup-DescribeFramework-response-FrameworkStatus)
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block).

## Import

Backup Framework can be imported using the `id` which corresponds to the name of the Backup Framework, e.g.,

```
$ terraform import aws_backup_framework.test <id>
```
