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

## Argument Reference

The following arguments are required:

* `account_access_type` - (Required) The type of account access for the workspace. Valid values are `CURRENT_ACCOUNT` and `ORGANIZATION`. If `ORGANIZATION` is specified, then `organizational_units` must also be present.
* `authentication_providers` - (Required) The authentication providers for the workspace. Valid values are `AWS_SSO`, `SAML`, or both.
* `permission_type` - (Required) The permission type of the workspace. If `SERVICE_MANAGED` is specified, the IAM roles and IAM policy attachments are generated automatically. If `CUSTOMER_MANAGED` is specified, the IAM roles and IAM policy attachments will not be created.

The following arguments are optional:

* `configuration` - (Optional) The configuration string for the workspace that you create. For more information about the format and configuration options available, see [Working in your Grafana workspace](https://docs.aws.amazon.com/grafana/latest/userguide/AMG-configure-workspace.html).
* `data_sources` - (Optional) The data sources for the workspace. Valid values are `AMAZON_OPENSEARCH_SERVICE`, `ATHENA`, `CLOUDWATCH`, `PROMETHEUS`, `REDSHIFT`, `SITEWISE`, `TIMESTREAM`, `XRAY`
* `description` - (Optional) The workspace description.
* `grafana_version` - (Optional) Specifies the version of Grafana to support in the new workspace. Supported values are `8.4`, `9.4` and `10.4`. If not specified, defaults to `9.4`.
* `name` - (Optional) The Grafana workspace name.
* `network_access_control` - (Optional) Configuration for network access to your workspace.See [Network Access Control](#network-access-control) below.
* `notification_destinations` - (Optional) The notification destinations. If a data source is specified here, Amazon Managed Grafana will create IAM roles and permissions needed to use these destinations. Must be set to `SNS`.
* `organization_role_name` - (Optional) The role name that the workspace uses to access resources through Amazon Organizations.
* `organizational_units` - (Optional) The Amazon Organizations organizational units that the workspace is authorized to use data sources from.
* `role_arn` - (Optional) The IAM role ARN that the workspace assumes.
* `stack_set_name` - (Optional) The AWS CloudFormation stack set name that provisions IAM roles to be used by the workspace.
* `tags` - (Optional) Key-value mapping of resource tags. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.
* `vpc_configuration` - (Optional) The configuration settings for an Amazon VPC that contains data sources for your Grafana workspace to connect to. See [VPC Configuration](#vpc-configuration) below.

### Network Access Control

* `prefix_list_ids` - (Required) - An array of prefix list IDs.
* `vpce_ids` - (Required) - An array of Amazon VPC endpoint IDs for the workspace. The only VPC endpoints that can be specified here are interface VPC endpoints for Grafana workspaces (using the com.amazonaws.[region].grafana-workspace service endpoint). Other VPC endpoints will be ignored.

### VPC Configuration

* `security_group_ids` - (Required) - The list of Amazon EC2 security group IDs attached to the Amazon VPC for your Grafana workspace to connect.
* `subnet_ids` - (Required) - The list of Amazon EC2 subnet IDs created in the Amazon VPC for your Grafana workspace to connect.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - The Amazon Resource Name (ARN) of the Grafana workspace.
* `endpoint` - The endpoint of the Grafana workspace.
* `grafana_version` - The version of Grafana running on the workspace.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

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
