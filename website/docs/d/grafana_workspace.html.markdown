---
subcategory: "Managed Grafana"
layout: "aws"
page_title: "AWS: aws_grafana_workspace"
description: |-
  Gets information on an Amazon Managed Grafana workspace.
---

# Data Source: aws_grafana_workspace

Provides an Amazon Managed Grafana workspace data source.

## Example Usage

### Basic configuration

```terraform
data "aws_grafana_workspace" "example" {
  workspace_id = "g-2054c75a02"
}
```

## Argument Reference

The following arguments are required:

* `workspace_id` - (Required) The Grafana workspace ID.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `account_access_type` - (Required) The type of account access for the workspace. Valid values are `CURRENT_ACCOUNT` and `ORGANIZATION`. If `ORGANIZATION` is specified, then `organizational_units` must also be present.
* `authentication_providers` - (Required) The authentication providers for the workspace. Valid values are `AWS_SSO`, `SAML`, or both.
* `arn` - The Amazon Resource Name (ARN) of the Grafana workspace.
* `created_date` - The creation date of the Grafana workspace.
* `data_sources` - The data sources for the workspace.
* `description` - The workspace description.
* `endpoint` - The endpoint of the Grafana workspace.
* `grafana_version` - The version of Grafana running on the workspace.
* `last_updated_date` - The last updated date of the Grafana workspace.
* `name` - The Grafana workspace name.
* `notification_destinations` - The notification destinations.
* `organization_role_name` - The role name that the workspace uses to access resources through Amazon Organizations.
* `organizational_units` - The Amazon Organizations organizational units that the workspace is authorized to use data sources from.
* `permission_type` - The permission type of the workspace.
* `role_arn` - The IAM role ARN that the workspace assumes.
* `stack_set_name` - The AWS CloudFormation stack set name that provisions IAM roles to be used by the workspace.
* `status` - The status of the Grafana workspace.
* `tags` - The tags assigned to the resource
