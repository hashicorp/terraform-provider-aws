---
subcategory: "Grafana"
layout: "aws"
page_title: "AWS: aws_grafana_workspace"
description: |-
Gents information on an Amazon Managed Grafana workspace.
---

# Resource: aws_grafana_workspace

Provides an Amazon Managed Grafana workspace resource.

## Example Usage

### Basic configuration

```terraform
data "aws_grafana_workspace" "example" {
  id = "g-2054c75a02"
}
```

## Argument Reference

The following arguments are required:

* `id` - (Required) The Grafana workspace ID.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:
* `account_access_type` - (Required) The type of account access for the workspace. Valid values are `CURRENT_ACCOUNT` and `ORGANIZATION`. If `ORGANIZATION` is specified, then `organizational_units` must also be present.
* `authentication_providers` - (Required) The authentication providers for the workspace. Valid values are `AWS_SSO`, `SAML`, or both.
* `arn` - The Amazon Resource Name (ARN) of the Grafana workspace.
* `status` - The status of the Grafana workspace. One of `ACTIVE`, `CREATING`, `DELETING`, `FAILED`, `UPDATING`, `UPGRADING`, `DELETION_FAILED`, `CREATION_FAILED`, `UPDATE_FAILED`, `UPGRADE_FAILED`, `LICENSE_REMOVAL_FAILED`.
* `created_date` - The creation date of the Grafana workspace.
* `last_updated_date` - The last updated date of the Grafana workspace.
* `endpoint` - The endpoint of the Grafana workspace.
* `grafana_version` - The version of Grafana running on the workspace.
* `organization_role_name` - (Optional) The role name that the workspace uses to access resources through Amazon Organizations.
* `permission_type` - (Optional) The permission type of the workspace. If `SERVICE_MANAGED` is specified, the IAM roles and IAM policy attachments are generated automatically. If `CUSTOMER_MANAGED` is specified, the IAM roles and IAM policy attachments will not be created.
* `stack_set_name` - (Optional) The AWS CloudFormation stack set name that provisions IAM roles to be used by the workspace.
* `data_sources` - (Optional) The data sources for the workspace. Valid values are `AMAZON_OPENSEARCH_SERVICE`, `CLOUDWATCH`, `PROMETHEUS`, `XRAY`, `TIMESTREAM`, `SITEWISE`.
* `description` - (Optional) The workspace description.
* `name` - (Optional) The Grafana workspace name.
* `notification_destinations` - (Optional) The notification destinations. If a data source is specified here, Amazon Managed Grafana will create IAM roles and permissions needed to use these destinations. Must be set to `SNS`.
* `organizational_units` - (Optional) The Amazon Organizations organizational units that the workspace is authorized to use data sources from.
* `role_arn` - (Optional) The IAM role ARN that the workspace assumes.