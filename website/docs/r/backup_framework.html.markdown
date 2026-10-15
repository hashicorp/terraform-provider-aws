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

  control {
    name = "BACKUP_RESOURCES_PROTECTED_BY_BACKUP_VAULT_LOCK"

    input_parameter {
      name  = "maxRetentionDays"
      value = "100"
    }

    input_parameter {
      name  = "minRetentionDays"
      value = "1"
    }

    scope {
      compliance_resource_types = [
        "EBS"
      ]
    }
  }

  control {
    name = "BACKUP_LAST_RECOVERY_POINT_CREATED"

    input_parameter {
      name  = "recoveryPointAgeUnit"
      value = "days"
    }

    input_parameter {
      name  = "recoveryPointAgeValue"
      value = "1"
    }

    scope {
      compliance_resource_types = [
        "EBS"
      ]
    }
  }

  tags = {
    "Name" = "Example Framework"
  }
}
```

## Argument Reference

This resource supports the following arguments:

* `control` - (Required) One or more control blocks that make up the framework. Each control in the list has a name, input parameters, and scope. Detailed below.
* `description` - (Optional) Description of the framework with a maximum of 1,024 characters
* `name` - (Required) Unique name of the framework. The name must be between 1 and 256 characters, starting with a letter, and consisting of letters, numbers, and underscores.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `tags` - (Optional) Metadata that you can assign to help organize the frameworks you create. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

### `control` Block

`control` has the following attributes:

* `input_parameter` - (Optional) One or more input parameter blocks. An example of a control with two parameters is: "backup plan frequency is at least daily and the retention period is at least 1 year". The first parameter is daily. The second parameter is 1 year. Detailed below.
* `name` - (Required) Name of a control. This name is between 1 and 256 characters.
* `scope` - (Optional) Scope of a control. The control scope defines what the control will evaluate. Three examples of control scopes are: a specific backup plan, all backup plans with a specific tag, or all backup plans. Detailed below.

### `input_parameter` Block

`input_parameter` has the following attributes:

* `name` - (Optional) Name of a parameter, for example, BackupPlanFrequency.
* `value` - (Optional) Value of parameter, for example, hourly.

### `scope` Block

`scope` has the following attributes:

* `compliance_resource_ids` - (Optional) ID of the only AWS resource that you want your control scope to contain. Minimum number of 1 item. Maximum number of 100 items.
* `compliance_resource_types` - (Optional) Whether the control scope includes one or more types of resources, such as EFS or RDS.
* `tags` - (Optional) Tag key-value pair applied to those AWS resources that you want to trigger an evaluation for a rule. A maximum of one key-value pair can be provided.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the backup framework.
* `creation_time` - Date and time that a framework is created, in Unix format and Coordinated Universal Time (UTC).
* `deployment_status` - Deployment status of a framework. The statuses are: `CREATE_IN_PROGRESS` | `UPDATE_IN_PROGRESS` | `DELETE_IN_PROGRESS` | `COMPLETED` | `FAILED`.
* `id` - ID of the backup framework.
* `status` - Framework consists of one or more controls. Each control governs a resource, such as backup plans, backup selections, backup vaults, or recovery points. You can also turn AWS Config recording on or off for each resource. For more information refer to the [AWS documentation for Framework Status](https://docs.aws.amazon.com/aws-backup/latest/devguide/API_DescribeFramework.html#Backup-DescribeFramework-response-FrameworkStatus)
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `2m`)
* `update` - (Default `2m`)
* `delete` - (Default `2m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Backup Framework using the `id` which corresponds to the name of the Backup Framework. For example:

```terraform
import {
  to = aws_backup_framework.test
  id = "<id>"
}
```

Using `terraform import`, import Backup Framework using the `id` which corresponds to the name of the Backup Framework. For example:

```console
% terraform import aws_backup_framework.test <id>
```
