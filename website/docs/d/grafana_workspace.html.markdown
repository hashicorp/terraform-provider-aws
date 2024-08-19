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

* `workspace_id` - (Required) Grafana workspace ID.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `account_access_type` - (Required) Type of account access for the workspace. Valid values are `CURRENT_ACCOUNT` and `ORGANIZATION`. If `ORGANIZATION` is specified, then `organizational_units` must also be present.
* `authentication_providers` - (Required) Authentication providers for the workspace. Valid values are `AWS_SSO`, `SAML`, or both.
* `arn` - ARN of the Grafana workspace.
* `created_date` - Creation date of the Grafana workspace.
* `data_sources` - Data sources for the workspace.
* `description` - Workspace description.
* `endpoint` - Endpoint of the Grafana workspace.
* `grafana_version` - Version of Grafana running on the workspace.
* `last_updated_date` - Last updated date of the Grafana workspace.
* `name` - Grafana workspace name.
* `notification_destinations` - The notification destinations.
* `organization_role_name` - The role name that the workspace uses to access resources through Amazon Organizations.
* `organizational_units` - The Amazon Organizations organizational units that the workspace is authorized to use data sources from.
* `permission_type` - Permission type of the workspace.
* `role_arn` - IAM role ARN that the workspace assumes.
* `stack_set_name` - AWS CloudFormation stack set name that provisions IAM roles to be used by the workspace.
* `status` - Status of the Grafana workspace.
* `tags` - Tags assigned to the resource
