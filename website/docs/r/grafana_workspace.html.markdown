---
subcategory: "Managed Grafana"
layout: "aws"
page_title: "AWS: aws_grafana_workspace"
description: |-
  Provides an Amazon Managed Grafana workspace resource.
---

# Resource: aws_grafana_workspace

Provides an Amazon Managed Grafana workspace resource.

## Example Usage

### Basic configuration

```terraform
resource "aws_grafana_workspace" "example" {
  account_access_type      = "CURRENT_ACCOUNT"
  authentication_providers = ["SAML"]
  permission_type          = "SERVICE_MANAGED"
  role_arn                 = aws_iam_role.assume.arn
}

resource "aws_iam_role" "assume" {
  name = "grafana-assume"
  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRole"
        Effect = "Allow"
        Sid    = ""
        Principal = {
          Service = "grafana.amazonaws.com"
        }
      },
    ]
  })
}
```

### Workspace configuration options

```terraform
resource "aws_grafana_workspace" "example" {
  account_access_type      = "CURRENT_ACCOUNT"
  authentication_providers = ["SAML"]
  permission_type          = "SERVICE_MANAGED"
  role_arn                 = aws_iam_role.assume.arn

  configuration = jsonencode(
    {
      "plugins" = {
        "pluginAdminEnabled" = true
      },
      "unifiedAlerting" = {
        "enabled" = false
      }
    }
  )
}
```

The optional argument `configuration` is a JSON string that disables the unified `Grafana Alerting` (Grafana version 10 or newer) and enables `Plugin Management` (Grafana version 9 or newer) on the Grafana Workspaces.

For more information about using Grafana alerting, and the effects of turning it on or off, see [Alerts in Grafana version 10](https://docs.aws.amazon.com/grafana/latest/userguide/v10-alerts.html).

## Argument Reference

The following arguments are required:

* `account_access_type` - (Required) Type of account access for the workspace. Valid values are `CURRENT_ACCOUNT` and `ORGANIZATION`. If `ORGANIZATION` is specified, then `organizational_units` must also be present.
* `authentication_providers` - (Required) Authentication providers for the workspace. Valid values are `AWS_SSO`, `SAML`, or both.
* `permission_type` - (Required) Permission type of the workspace. If `SERVICE_MANAGED` is specified, the IAM roles and IAM policy attachments are generated automatically. If `CUSTOMER_MANAGED` is specified, the IAM roles and IAM policy attachments will not be created.

The following arguments are optional:

* `configuration` - (Optional) Configuration string for the workspace that you create. For more information about the format and configuration options available, see [Working in your Grafana workspace](https://docs.aws.amazon.com/grafana/latest/userguide/AMG-configure-workspace.html).
* `data_sources` - (Optional) Data sources for the workspace. Valid values are `AMAZON_OPENSEARCH_SERVICE`, `ATHENA`, `CLOUDWATCH`, `PROMETHEUS`, `REDSHIFT`, `SITEWISE`, `TIMESTREAM`, `TWINMAKER`, `XRAY`
* `description` - (Optional) Workspace description.
* `grafana_version` - (Optional) Version of Grafana to support in the new workspace. Supported values are `9.4`, `10.4`, `12.4` and `13.2`. If not specified, defaults to the latest version.
* `kms_key_id` - (Optional) ARN of the AWS KMS key for encrypting workspace data.
* `name` - (Optional) Grafana workspace name.
* `network_access_control` - (Optional) Configuration for network access to your workspace. See [`network_access_control` Block](#network_access_control-block) below.
* `notification_destinations` - (Optional) Notification destinations. If a data source is specified here, Amazon Managed Grafana will create IAM roles and permissions needed to use these destinations. Must be set to `SNS`.
* `organization_role_name` - (Optional) Role name that the workspace uses to access resources through Amazon Organizations.
* `organizational_units` - (Optional) Amazon Organizations organizational units that the workspace is authorized to use data sources from.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `role_arn` - (Optional) IAM role ARN that the workspace assumes.
* `stack_set_name` - (Optional) AWS CloudFormation stack set name that provisions IAM roles to be used by the workspace.
* `tags` - (Optional) Key-value mapping of resource tags. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.
* `vpc_configuration` - (Optional) Configuration settings for an Amazon VPC that contains data sources for your Grafana workspace to connect to. See [`vpc_configuration` Block](#vpc_configuration-block) below.

### `network_access_control` Block

* `prefix_list_ids` - (Required) - An array of prefix list IDs.
* `vpce_ids` - (Required) - An array of Amazon VPC endpoint IDs for the workspace. The only VPC endpoints that can be specified here are interface VPC endpoints for Grafana workspaces (using the com.amazonaws.[region].grafana-workspace service endpoint). Other VPC endpoints will be ignored.

### `vpc_configuration` Block

* `security_group_ids` - (Required) - The list of Amazon EC2 security group IDs attached to the Amazon VPC for your Grafana workspace to connect.
* `subnet_ids` - (Required) - The list of Amazon EC2 subnet IDs created in the Amazon VPC for your Grafana workspace to connect.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the Grafana workspace.
* `endpoint` - Endpoint of the Grafana workspace.
* `grafana_version` - Version of Grafana running on the workspace.
* `saml_configuration_status` - Status of the SAML configuration.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `30m`)
* `update` - (Default `30m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Grafana Workspace using the workspace's `id`. For example:

```terraform
import {
  to = aws_grafana_workspace.example
  id = "g-2054c75a02"
}
```

Using `terraform import`, import Grafana Workspace using the workspace's `id`. For example:

```console
% terraform import aws_grafana_workspace.example g-2054c75a02
```
