---
subcategory: "CodeCatalyst"
layout: "aws"
page_title: "AWS: aws_codecatalyst_dev_environment"
description: |-
  Terraform data source for managing an AWS CodeCatalyst Dev Environment.
---
# Data Source: aws_codecatalyst_dev_environment

Terraform data source for managing an AWS CodeCatalyst Dev Environment.

## Example Usage

### Basic Usage

```terraform
data "aws_codecatalyst_dev_environment" "example" {
  space_name   = "myspace"
  project_name = "myproject"
  env_id       = aws_codecatalyst_dev_environment.example.id
}
```

## Argument Reference

The following arguments are required:

* `env_id` - - (Required) The system-generated unique ID of the Dev Environment for which you want to view information. To retrieve a list of Dev Environment IDs, use [ListDevEnvironments](https://docs.aws.amazon.com/codecatalyst/latest/APIReference/API_ListDevEnvironments.html).
* `project_name` - (Required) The name of the project in the space.
* `space_name` - (Required) The name of the space.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `alias` - The user-specified alias for the Dev Environment.
* `creator_id` - The system-generated unique ID of the user who created the Dev Environment.
* `ides` - Information about the integrated development environment (IDE) configured for a Dev Environment.
* `inactivity_timeout_minutes` - The amount of time the Dev Environment will run without any activity detected before stopping, in minutes. Only whole integers are allowed. Dev Environments consume compute minutes when running.
* `instance_type` - The Amazon EC2 instace type to use for the Dev Environment.
* `last_updated_time` - The time when the Dev Environment was last updated, in coordinated universal time (UTC) timestamp format as specified in [RFC 3339](https://www.rfc-editor.org/rfc/rfc3339#section-5.6).
* `persistent_storage` - Information about the amount of storage allocated to the Dev Environment.
* `repositories` - The source repository that contains the branch to clone into the Dev Environment.
* `status` - The current status of the Dev Environment. From: PENDING | RUNNING | STARTING | STOPPING | STOPPED | FAILED | DELETING | DELETED.
* `status_reason` - The reason for the status.
