---
subcategory: "Backup"
layout: "aws"
page_title: "AWS: aws_backup_framework"
description: |-
  Provides details about an AWS Backup Framework.
---

# Data Source: aws_backup_framework

Use this data source to get information on an existing backup framework.

## Example Usage

```terraform
data "aws_backup_framework" "example" {
  name = "tf_example_backup_framework_name"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The backup framework name.

## Attributes Reference

In addition to the arguments above, the following attributes are exported:

* `arn` - The ARN of the backup framework.
* `control` - One or more control blocks that make up the framework. Each control in the list has a name, input parameters, and scope. Detailed below.
* `creation_time` - The date and time that a framework is created, in Unix format and Coordinated Universal Time (UTC).
* `deployment_status` - The deployment status of a framework. The statuses are: `CREATE_IN_PROGRESS` | `UPDATE_IN_PROGRESS` | `DELETE_IN_PROGRESS` | `COMPLETED`| `FAILED`.
* `description` - The description of the framework.
* `id` - The id of the framework.
* `status` - A framework consists of one or more controls. Each control governs a resource, such as backup plans, backup selections, backup vaults, or recovery points. You can also turn AWS Config recording on or off for each resource. The statuses are: `ACTIVE`, `PARTIALLY_ACTIVE`, `INACTIVE`, `UNAVAILABLE`. For more information refer to the [AWS documentation for Framework Status](https://docs.aws.amazon.com/aws-backup/latest/devguide/API_DescribeFramework.html#Backup-DescribeFramework-response-FrameworkStatus)
* `tags` - Metadata that helps organize the frameworks you create.

### Control Attributes
For **control** the following attributes are supported:

* `input_parameter` - One or more input parameter blocks. An example of a control with two parameters is: "backup plan frequency is at least daily and the retention period is at least 1 year". The first parameter is daily. The second parameter is 1 year. Detailed below.
* `name` - The name of a control.
* `scope` - The scope of a control. The control scope defines what the control will evaluate. Three examples of control scopes are: a specific backup plan, all backup plans with a specific tag, or all backup plans. Detailed below.

### Input Parameter Attributes
For **input_parameter** the following attributes are supported:

* `name` - The name of a parameter, for example, BackupPlanFrequency.
* `value` - The value of parameter, for example, hourly.

### Scope Attributes
For **scope** the following attributes are supported:

* `compliance_resource_ids` - The ID of the only AWS resource that you want your control scope to contain.
* `compliance_resource_types` - Describes whether the control scope includes one or more types of resources, such as EFS or RDS.
* `tags` - The tag key-value pair applied to those AWS resources that you want to trigger an evaluation for a rule. A maximum of one key-value pair can be provided.
