---
subcategory: ""
layout: "aws"
page_title: "Terraform AWS Provider Version 5 Upgrade Guide"
description: |-
  Terraform AWS Provider Version 5 Upgrade Guide
---

# Terraform AWS Provider Version 5 Upgrade Guide

Version 5.0.0 of the AWS provider for Terraform is a major release and includes some changes that you will need to consider when upgrading. We intend this guide to help with that process and focus only on changes from version 4.X to version 5.0.0. See the [Version 4 Upgrade Guide](/docs/providers/aws/guides/version-4-upgrade.html) for information about upgrading from 3.X to version 4.0.0.

We previously marked most of the changes we outline in this guide as deprecated in the Terraform plan/apply output throughout previous provider releases. You can find these changes, including deprecation notices, in the [Terraform AWS Provider CHANGELOG](https://github.com/hashicorp/terraform-provider-aws/blob/main/CHANGELOG.md).

Upgrade topics:

<!-- TOC depthFrom:2 depthTo:2 -->

- [Provider Version Configuration](#provider-version-configuration)
- [Provider Arguments](#provider-arguments)
- [Default Tags](#default-tags)
- [Data Source: aws_api_gateway_rest_api](#data-source-aws_api_gateway_rest_api)
- [Data Source: aws_connect_hours_of_operation](#data-source-aws_connect_hours_of_operation)
- [Data Source: aws_identitystore_group](#data-source-aws_identitystore_group)
- [Data Source: aws_identitystore_user](#data-source-aws_identitystore_user)
- [Data Source: aws_quicksight_data_set](#data-source-aws_quicksight_data_set)
- [Data Source: aws_redshift_service_account](#data-source-aws_redshift_service_account)
- [Data Source: aws_secretsmanager_secret](#data-source-aws_secretsmanager_secret)
- [Data Source: aws_service_discovery_service](#data-source-aws_service_discovery_service)
- [Data Source: aws_subnet_ids](#data-source-aws_subnet_ids)
- [Resource: aws_acmpca_certificate_authority](#resource-aws_acmpca_certificate_authority)
- [Resource: aws_api_gateway_rest_api](#resource-aws_api_gateway_rest_api)
- [Resource: aws_autoscaling_group](#resource-aws_autoscaling_group)
- [Resource: aws_budgets_budget](#resource-aws_budgets_budget)
- [Resource: aws_ce_anomaly_subscription](#resource-aws_ce_anomaly_subscription)
- [Resource: aws_cloudwatch_event_target](#resource-aws_cloudwatch_event_target)
- [Resource: aws_codebuild_project](#resource-aws_codebuild_project)
- [Resource: aws_connect_hours_of_operation](#resource-aws_connect_hours_of_operation)
- [Resource: aws_connect_queue](#resource-aws_connect_queue)
- [Resource: aws_connect_routing_profile](#resource-aws_connect_routing_profile)
- [Resource: aws_docdb_cluster](#resource-aws_docdb_cluster)
- [Resource: aws_dx_gateway_association](#resource-aws_dx_gateway_association)
- [Resource: aws_ec2_client_vpn_endpoint](#resource-aws_ec2_client_vpn_endpoint)
- [Resource: aws_ec2_client_vpn_network_association](#resource-aws_ec2_client_vpn_network_association)
- [Resource: aws_ecs_cluster](#resource-aws_ecs_cluster)
- [Resource: aws_eks_addon](#resource-aws_eks_addon)
- [Resource: aws_lightsail_instance](#resource-aws_lightsail_instance)
- [Resource: aws_msk_cluster](#resource-aws_msk_cluster)
- [Resource: aws_neptune_cluster](#resource-aws_neptune_cluster)
- [Resource: aws_rds_cluster](#resource-aws_rds_cluster)
- [Resource: aws_secretsmanager_secret](#resource-aws_secretsmanager_secret)
- [Resource: aws_wafv2_web_acl](#resource-aws_wafv2_web_acl)
- [Resource: aws_wafv2_web_acl_logging_configuration](#resource-aws_wafv2_web_acl_logging_configuration)

<!-- /TOC -->

Additional Topics:

<!-- TOC depthFrom:2 depthTo:2 -->

- [EC2-Classic Retirement](#ec2-classic-retirement)
- [Macie Classic Retirement](#macie-classic-retirement)

<!-- /TOC -->

## Provider Version Configuration

-> Before upgrading to version 5.0.0, upgrade to the most recent 4.X version of the provider and ensure that your environment successfully runs [`terraform plan`](https://www.terraform.io/docs/commands/plan.html). You should not see changes you don't expect or deprecation notices.

Use [version constraints when configuring Terraform providers](https://www.terraform.io/docs/configuration/providers.html#provider-versions). If you are following that recommendation, update the version constraints in your Terraform configuration and run [`terraform init -upgrade`](https://www.terraform.io/docs/commands/init.html) to download the new version.

For example, given this previous configuration:

```terraform
terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 4.65"
    }
  }
}

provider "aws" {
  # Configuration options
}
```

Update to the latest 5.X version:

```terraform
terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
  }
}

provider "aws" {
  # Configuration options
}
```

## Provider Arguments

Version 5.0.0 removes these `provider` arguments:

* `assume_role.duration_seconds` - Use `assume_role.duration` instead
* `assume_role_with_web_identity.duration_seconds` - Use `assume_role_with_web_identity.duration` instead
* `s3_force_path_style` - Use `s3_use_path_style` instead
* `shared_credentials_file` - Use `shared_credentials_files` instead
* `skip_get_ec2_platforms` - Removed following the retirement of EC2-Classic

## Resource: aws_acmpca_certificate_authority

The `status` attribute is superfluous and sometimes incorrect. It has been removed.

## Resource: aws_api_gateway_rest_api

The `minimum_compression_size` attribute is now a String type, allowing it to be computed when set via the `body` attribute. Valid values remain the same.

## Resource: aws_autoscaling_group

The `tags` attribute has been removed. Use the `tag` attribute instead. For use cases requiring dynamic tags, see the [Dynamic Tagging example](../r/autoscaling_group.html.markdown#dynamic-tagging).

## Resource: aws_budgets_budget

The `cost_filters` attribute has been removed.

## Resource: aws_ce_anomaly_subscription

The `threshold` attribute has been removed.

## Resource: aws_cloudwatch_event_target

The `ecs_target.propagate_tags` attribute now has no default value. If no value is specified, the tags are not propagated.

## Resource: aws_codebuild_project

The `secondary_sources.auth` and `source.auth` attributes have been removed.

## Resource: aws_connect_hours_of_operation

The `hours_of_operation_arn` attribute has been removed.

## Resource: aws_connect_queue

The `quick_connect_ids_associated` attribute has been removed.

## Resource: aws_connect_routing_profile

The `queue_configs_associated` attribute has been removed.

## Resource: aws_docdb_cluster

Changes to the `snapshot_identifier` attribute will now correctly force re-creation of the resource. Previously, changing this attribute would result in a successful apply, but without the cluster being restored (only the resource state was changed). This change brings behavior of the cluster `snapshot_identifier` attribute into alignment with other RDS resources, such as `aws_db_instance`.

Automated snapshots **should not** be used for this attribute, unless from a different cluster. Automated snapshots are deleted as part of cluster destruction when the resource is replaced.

## Resource: aws_dx_gateway_association

The `vpn_gateway_id` attribute has been deprecated. All configurations using `vpn_gateway_id` should be updated to use the `associated_gateway_id` attribute instead.

## Resource: aws_ec2_client_vpn_endpoint

The `security_groups` and `status` attributes have been removed.

## Resource: aws_ec2_client_vpn_network_association

The `status` attribute has been removed.

## Resource: aws_ecs_cluster

The `capacity_providers` and `default_capacity_provider_strategy` attributes have been removed.

## Resource: aws_eks_addon

The `resolve_conflicts` argument has been deprecated. Use the `resolve_conflicts_on_create` and/or `resolve_conflicts_on_update` arguments instead.

## Resource: aws_lightsail_instance

The `ipv6_address` attribute has been removed.

## Resource: aws_msk_cluster

The `broker_node_group_info.ebs_volume_size` attribute has been removed.

## Resource: aws_neptune_cluster

Changes to the `snapshot_identifier` attribute will now correctly force re-creation of the resource. Previously, changing this attribute would result in a successful apply, but without the cluster being restored (only the resource state was changed). This change brings behavior of the cluster `snapshot_identifier` attribute into alignment with other RDS resources, such as `aws_db_instance`.

Automated snapshots **should not** be used for this attribute, unless from a different cluster. Automated snapshots are deleted as part of cluster destruction when the resource is replaced.

## Resource: aws_rds_cluster

Changes to the `snapshot_identifier` attribute will now correctly force re-creation of the resource. Previously, changing this attribute would result in a successful apply, but without the cluster being restored (only the resource state was changed). This change brings behavior of the cluster `snapshot_identifier` attribute into alignment with other RDS resources, such as `aws_db_instance`.

Automated snapshots **should not** be used for this attribute, unless from a different cluster. Automated snapshots are deleted as part of cluster destruction when the resource is replaced.

## Resource: aws_secretsmanager_secret

The `rotation_enabled`, `rotation_lambda_arn` and `rotation_rules` attributes have been removed.

## Resource: aws_wafv2_web_acl

The `statement.managed_rule_group_statement.excluded_rule` and `statement.rule_group_reference_statement.excluded_rule` attributes have been removed.

The `statement.rule_group_reference_statement.rule_action_override` attribute has been added.

## Resource: aws_wafv2_web_acl_logging_configuration

The `redacted_fields.all_query_arguments`, `redacted_fields.body` and `redacted_fields.single_query_argument` attributes have been removed.

## Data Source: aws_api_gateway_rest_api

The `minimum_compression_size` attribute is now a String type, allowing it to be computed when set via the `body` attribute.

## Data Source: aws_connect_hours_of_operation

The `hours_of_operation_arn` attribute has been removed.

## Data Source: aws_identitystore_group

The `filter` argument has been removed.

## Data Source: aws_identitystore_user

The `filter` argument has been removed.

## Data Source: aws_quicksight_data_set

The `tags_all` attribute has been deprecated and will be removed in a future version.

## Data Source: aws_redshift_service_account

[AWS document](https://docs.aws.amazon.com/redshift/latest/mgmt/db-auditing.html#db-auditing-bucket-permissions) that [a service principal name](https://docs.aws.amazon.com/IAM/latest/UserGuide/reference_policies_elements_principal.html#principal-services) be used instead of AWS account ID in any relevant IAM policy.
The [`aws_redshift_service_account`](/docs/providers/aws/d/redshift_service_account.html) data source should now be considered deprecated and will be removed in a future version.

## Data Source: aws_service_discovery_service

The `tags_all` attribute has been deprecated and will be removed in a future version.

## Data Source: aws_secretsmanager_secret

The `rotation_enabled`, `rotation_lambda_arn` and `rotation_rules` attributes have been removed.

## Data Source: aws_subnet_ids

The `aws_subnet_ids` data source has been removed. Use the [`aws_subnets`](/docs/providers/aws/d/subnets.html) data source instead.

## Default Tags

The following enhancements are included:

* Duplicate `default_tags` can now be included and will be overwritten by resource `tags`.
* Zero value tags, `""`, can now be included in both `default_tags` and resource `tags`.
* Tags can now be `computed`.

## EC2-Classic Retirement

Following the retirement of EC2-Classic a number of resources and attributes have been removed.

* The `aws_db_security_group` resource has been removed
* The `aws_elasticache_security_group` resource has been removed
* The `aws_redshift_security_group` resource has been removed
* The [`aws_db_instance`](/docs/providers/aws/r/db_instance.html) resource's `security_group_names` argument has been removed
* The [`aws_elasticache_cluster`](/docs/providers/aws/r/elasticache_cluster.html) resource's `security_group_names` argument has been removed
* The [`aws_redshift_cluster`](/docs/providers/aws/r/redshift_cluster.html) resource's `cluster_security_groups` argument has been removed
* The [`aws_launch_configuration`](/docs/providers/aws/r/launch_configuration.html) resource's `vpc_classic_link_id` and `vpc_classic_link_security_groups` arguments have been removed
* The [`aws_vpc`](/docs/providers/aws/r/vpc.html) resource's `enable_classiclink` and `enable_classiclink_dns_support` arguments have been removed
* The [`aws_default_vpc`](/docs/providers/aws/r/default_vpc.html) resource's `enable_classiclink` and `enable_classiclink_dns_support` arguments have been removed
* The [`aws_vpc_peering_connection`](/docs/providers/aws/r/vpc_peering_connection.html) resource's `allow_classic_link_to_remote_vpc` and `allow_vpc_to_remote_classic_link` arguments have been removed
* The [`aws_vpc_peering_connection_accepter`](/docs/providers/aws/r/vpc_peering_connection_accepter.html) resource's `allow_classic_link_to_remote_vpc` and `allow_vpc_to_remote_classic_link` arguments have been removed
* The [`aws_vpc_peering_connection_options`](/docs/providers/aws/r/vpc_peering_connection_options.html) resource's `allow_classic_link_to_remote_vpc` and `allow_vpc_to_remote_classic_link` arguments have been removed
* The [`aws_db_instance`](/docs/providers/aws/d/db_instance.html) data source's `db_security_groups` attribute has been removed
* The [`aws_elasticache_cluster`](/docs/providers/aws/d/elasticache_cluster.html) data source's `security_group_names` attribute has been removed
* The [`aws_redshift_cluster`](/docs/providers/aws/d/redshift_cluster.html) data source's `cluster_security_groups` attribute has been removed
* The [`aws_launch_configuration`](/docs/providers/aws/d/launch_configuration.html) data source's `vpc_classic_link_id` and `vpc_classic_link_security_groups` attributes have been removed

## Macie Classic Retirement

Following the retirement of Amazon Macie Classic a couple of resources have been removed.

* The `aws_macie_member_account_association` resource has been removed
* The `aws_macie_s3_bucket_association` resource has been removed
