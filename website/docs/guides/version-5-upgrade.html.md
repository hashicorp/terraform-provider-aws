---
subcategory: ""
layout: "aws"
page_title: "Terraform AWS Provider Version 5 Upgrade Guide"
description: |-
  Terraform AWS Provider Version 5 Upgrade Guide
---

# Terraform AWS Provider Version 5 Upgrade Guide

Version 5.0.0 of the AWS provider for Terraform is a major release and includes changes that you need to consider when upgrading. This guide will help with that process and focuses only on changes from version 4.x to version 5.0.0. See the [Version 4 Upgrade Guide](/docs/providers/aws/guides/version-4-upgrade.html) for information on upgrading from 3.x to version 4.0.0.

Upgrade topics:

<!-- TOC depthFrom:2 depthTo:2 -->

- [Provider Version Configuration](#provider-version-configuration)
- [Provider Arguments](#provider-arguments)
- [Default Tags](#default-tags)
- [EC2 Classic Retirement](#ec2-classic-retirement)
- [Macie Classic Retirement](#macie-classic-retirement)
- [resource/aws_acmpca_certificate_authority](#resourceaws_acmpca_certificate_authority)
- [resource/aws_api_gateway_rest_api](#resourceaws_api_gateway_rest_api)
- [resource/aws_autoscaling_group](#resourceaws_autoscaling_group)
- [resource/aws_budgets_budget](#resourceaws_budgets_budget)
- [resource/aws_ce_anomaly_subscription](#resourceaws_ce_anomaly_subscription)
- [resource/aws_cloudwatch_event_target](#resourceaws_cloudwatch_event_target)
- [resource/aws_codebuild_project](#resourceaws_codebuild_project)
- [resource/aws_connect_hours_of_operation](#resourceaws_connect_hours_of_operation)
- [resource/aws_connect_queue](#resourceaws_connect_queue)
- [resource/aws_connect_routing_profile](#resourceaws_connect_routing_profile)
- [resource/aws_db_instance](#resourceaws_db_instance)
- [resource/aws_db_security_group](#resourceaws_db_security_group)
- [resource/aws_default_vpc](#resourceaws_default_vpc)
- [resource/aws_docdb_cluster](#resourceaws_docdb_cluster)
- [resource/aws_dx_gateway_association](#resourceaws_dx_gateway_association)
- [resource/aws_ec2_client_vpn_endpoint](#resourceaws_ec2_client_vpn_endpoint)
- [resource/aws_ec2_client_vpn_network_association](#resourceaws_ec2_client_vpn_network_association)
- [resource/aws_ecs_cluster](#resourceaws_ecs_cluster)
- [resource/aws_eks_addon](#resourceaws_eks_addon)
- [resource/aws_elasticache_cluster](#resourceaws_elasticache_cluster)
- [resource/aws_elasticache_security_group](#resourceaws_elasticache_security_group)
- [resource/aws_flow_log](#resourceaws_flow_log)
- [resource/aws_launch_configuration](#resourceaws_launch_configuration)
- [resource/aws_lightsail_instance](#resourceaws_lightsail_instance)
- [resource/aws_macie_member_account_association](#resourceaws_macie_member_account_association)
- [resource/aws_macie_s3_bucket_association](#resourceaws_macie_s3_bucket_association)
- [resource/aws_msk_cluster](#resourceaws_msk_cluster)
- [resource/aws_neptune_cluster](#resourceaws_neptune_cluster)
- [resource/aws_opensearch_domain](#resourceaws_opensearch_domain)
- [resource/aws_rds_cluster](#resourceaws_rds_cluster)
- [resource/aws_redshift_cluster](#resourceaws_redshift_cluster)
- [resource/aws_redshift_security_group](#resourceaws_redshift_security_group)
- [resource/aws_secretsmanager_secret](#resourceaws_secretsmanager_secret)
- [resource/aws_ssm_association](#resourceaws_ssm_association)
- [resource/aws_vpc](#resourceaws_vpc)
- [resource/aws_vpc_peering_connection](#resourceaws_vpc_peering_connection)
- [resource/aws_vpc_peering_connection_accepter](#resourceaws_vpc_peering_connection_accepter)
- [resource/aws_vpc_peering_connection_options](#resourceaws_vpc_peering_connection_options)
- [resource/aws_wafv2_web_acl](#resourceaws_wafv2_web_acl)
- [resource/aws_wafv2_web_acl_logging_configuration](#resourceaws_wafv2_web_acl_logging_configuration)
- [data-source/aws_api_gateway_rest_api](#data-sourceaws_api_gateway_rest_api)
- [data-source/aws_connect_hours_of_operation](#data-sourceaws_connect_hours_of_operation)
- [data-source/aws_db_instance](#data-sourceaws_db_instance)
- [data-source/aws_elasticache_cluster](#data-sourceaws_elasticache_cluster)
- [data-source/aws_identitystore_group](#data-sourceaws_identitystore_group)
- [data-source/aws_identitystore_user](#data-sourceaws_identitystore_user)
- [data-source/aws_launch_configuration](#data-sourceaws_launch_configuration)
- [data-source/aws_opensearch_domain](#data-sourceaws_opensearch_domain)
- [data-source/aws_quicksight_data_set](#data-sourceaws_quicksight_data_set)
- [data-source/aws_redshift_cluster](#data-sourceaws_redshift_cluster)
- [data-source/aws_redshift_service_account](#data-sourceaws_redshift_service_account)
- [data-source/aws_secretsmanager_secret](#data-sourceaws_secretsmanager_secret)
- [data-source/aws_service_discovery_service](#data-sourceaws_service_discovery_service)
- [data-source/aws_subnet_ids](#data-sourceaws_subnet_ids)

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
* `skip_get_ec2_platforms` - Removed following the retirement of EC2 Classic

## Default Tags

The following enhancements are included:

* Duplicate `default_tags` can now be included and will be overwritten by resource `tags`.
* Zero value tags, `""`, can now be included in both `default_tags` and resource `tags`.
* Tags can now be `computed`.

## EC2 Classic Retirement

Following the retirement of EC2 Classic, we removed a number of resources, arguments, and attributes. This list summarizes what we _removed_:

* `aws_db_security_group` resource
* `aws_elasticache_security_group` resource
* `aws_redshift_security_group` resource
* [`aws_db_instance`](/docs/providers/aws/r/db_instance.html) resource's `security_group_names` argument
* [`aws_elasticache_cluster`](/docs/providers/aws/r/elasticache_cluster.html) resource's `security_group_names` argument
* [`aws_redshift_cluster`](/docs/providers/aws/r/redshift_cluster.html) resource's `cluster_security_groups` argument
* [`aws_launch_configuration`](/docs/providers/aws/r/launch_configuration.html) resource's `vpc_classic_link_id` and `vpc_classic_link_security_groups` arguments
* [`aws_vpc`](/docs/providers/aws/r/vpc.html) resource's `enable_classiclink` and `enable_classiclink_dns_support` arguments
* [`aws_default_vpc`](/docs/providers/aws/r/default_vpc.html) resource's `enable_classiclink` and `enable_classiclink_dns_support` arguments
* [`aws_vpc_peering_connection`](/docs/providers/aws/r/vpc_peering_connection.html) resource's `allow_classic_link_to_remote_vpc` and `allow_vpc_to_remote_classic_link` arguments
* [`aws_vpc_peering_connection_accepter`](/docs/providers/aws/r/vpc_peering_connection_accepter.html) resource's `allow_classic_link_to_remote_vpc` and `allow_vpc_to_remote_classic_link` arguments
* [`aws_vpc_peering_connection_options`](/docs/providers/aws/r/vpc_peering_connection_options.html) resource's `allow_classic_link_to_remote_vpc` and `allow_vpc_to_remote_classic_link` arguments
* [`aws_db_instance`](/docs/providers/aws/d/db_instance.html) data source's `db_security_groups` attribute
* [`aws_elasticache_cluster`](/docs/providers/aws/d/elasticache_cluster.html) data source's `security_group_names` attribute
* [`aws_redshift_cluster`](/docs/providers/aws/d/redshift_cluster.html) data source's `cluster_security_groups` attribute
* [`aws_launch_configuration`](/docs/providers/aws/d/launch_configuration.html) data source's `vpc_classic_link_id` and `vpc_classic_link_security_groups` attributes

## Macie Classic Retirement

Following the retirement of Amazon Macie Classic, we removed these resources:

* `aws_macie_member_account_association`
* `aws_macie_s3_bucket_association`

## resource/aws_acmpca_certificate_authority

We removed the `status` attribute as it is superfluous and sometimes incorrect.

## resource/aws_api_gateway_rest_api

The `minimum_compression_size` attribute is now a String type, allowing it to be computed when set via the `body` attribute. Valid values remain the same.

## resource/aws_autoscaling_group

We removed the `tags` attribute. Use the `tag` attribute instead. For use cases requiring dynamic tags, see the [Dynamic Tagging example](../r/autoscaling_group.html.markdown#dynamic-tagging).

## resource/aws_budgets_budget

We removed the `cost_filters` attribute.

## resource/aws_ce_anomaly_subscription

We removed the `threshold` attribute.

## resource/aws_cloudwatch_event_target

The `ecs_target.propagate_tags` attribute now has no default value. If no value is specified, the tags are not propagated.

## resource/aws_codebuild_project

We removed the `secondary_sources.auth` and `source.auth` attributes.

## resource/aws_connect_hours_of_operation

We removed the `hours_of_operation_arn` attribute.

## resource/aws_connect_queue

We removed the `quick_connect_ids_associated` attribute.

## resource/aws_connect_routing_profile

We removed the `queue_configs_associated` attribute.

## resource/aws_db_instance

We removed the `db_security_groups` attribute as part of the EC2 Classic retirement.

## resource/aws_db_security_group

We removed this resource as part of the EC2 Classic retirement.

## resource/aws_default_vpc

We removed the `enable_classiclink` and `enable_classiclink_dns_support` arguments as part of the EC2 Classic retirement.

## resource/aws_docdb_cluster

Changes to the `snapshot_identifier` attribute will now correctly force re-creation of the resource. Previously, changing this attribute would result in a successful apply, but without the cluster being restored (only the resource state was changed). This change brings behavior of the cluster `snapshot_identifier` attribute into alignment with other RDS resources, such as `aws_db_instance`.

Automated snapshots **should not** be used for this attribute, unless from a different cluster. Automated snapshots are deleted as part of cluster destruction when the resource is replaced.

## resource/aws_dx_gateway_association

The `vpn_gateway_id` attribute has been deprecated. All configurations using `vpn_gateway_id` should be updated to use the `associated_gateway_id` attribute instead.

## resource/aws_ec2_client_vpn_endpoint

We removed the `security_groups` and `status` attributes.

## resource/aws_ec2_client_vpn_network_association

We removed the `status` attribute.

## resource/aws_ecs_cluster

We removed the `capacity_providers` and `default_capacity_provider_strategy` attributes.

## resource/aws_eks_addon

The `resolve_conflicts` argument has been deprecated. Use the `resolve_conflicts_on_create` and/or `resolve_conflicts_on_update` arguments instead.

## resource/aws_elasticache_cluster

We removed the `security_group_names` attribute as part of the EC2 Classic retirement.

## resource/aws_elasticache_security_group

We removed this resource as part of the EC2 Classic retirement.

## resource/aws_flow_log

The `log_group_name` attribute has been deprecated. All configurations using `log_group_name` should be updated to use the `log_destination` attribute instead.

## resource/aws_launch_configuration

We removed the `vpc_classic_link_id` and `vpc_classic_link_security_groups` arguments as part of the EC2 Classic retirement.

## resource/aws_lightsail_instance

We removed the `ipv6_address` attribute.

## resource/aws_macie_member_account_association

We removed this resource as part of the Macie Classic retirement.

## resource/aws_macie_s3_bucket_association

We removed this resource as part of the Macie Classic retirement.

## resource/aws_msk_cluster

We removed the `broker_node_group_info.ebs_volume_size` attribute.

## resource/aws_neptune_cluster

Changes to the `snapshot_identifier` attribute will now correctly force re-creation of the resource. Previously, changing this attribute would result in a successful apply, but without the cluster being restored (only the resource state was changed). This change brings behavior of the cluster `snapshot_identifier` attribute into alignment with other RDS resources, such as `aws_db_instance`.

Automated snapshots **should not** be used for this attribute, unless from a different cluster. Automated snapshots are deleted as part of cluster destruction when the resource is replaced.

## resource/aws_opensearch_domain

The `kibana_endpoint` attribute has been deprecated. All configurations using `kibana_endpoint` should be updated to use the `dashboard_endpoint` attribute instead.

## resource/aws_rds_cluster

Changes to the `snapshot_identifier` attribute will now correctly force re-creation of the resource. Previously, changing this attribute would result in a successful apply, but without the cluster being restored (only the resource state was changed). This change brings behavior of the cluster `snapshot_identifier` attribute into alignment with other RDS resources, such as `aws_db_instance`.

Automated snapshots **should not** be used for this attribute, unless from a different cluster. Automated snapshots are deleted as part of cluster destruction when the resource is replaced.

## resource/aws_redshift_cluster

We removed the `cluster_security_groups` attribute as part of the EC2 Classic retirement.

## resource/aws_redshift_security_group

We removed this resource as part of the EC2 Classic retirement.

## resource/aws_secretsmanager_secret

We removed the `rotation_enabled`, `rotation_lambda_arn` and `rotation_rules` attributes.

## resource/aws_ssm_association

The `instance_id` attribute has been deprecated. All configurations using `instance_id` should be updated to use the `targets` attribute instead.

## resource/aws_vpc

We removed the `enable_classiclink` and `enable_classiclink_dns_support` arguments as part of the EC2 Classic retirement.

## resource/aws_vpc_peering_connection

We removed the `allow_classic_link_to_remote_vpc` and `allow_vpc_to_remote_classic_link` arguments as part of the EC2 Classic retirement.

## resource/aws_vpc_peering_connection_accepter

We removed the `allow_classic_link_to_remote_vpc` and `allow_vpc_to_remote_classic_link` arguments as part of the EC2 Classic retirement.

## resource/aws_vpc_peering_connection_options

We removed the `allow_classic_link_to_remote_vpc` and `allow_vpc_to_remote_classic_link` arguments as part of the EC2 Classic retirement.

## resource/aws_wafv2_web_acl

We removed the `statement.managed_rule_group_statement.excluded_rule` and `statement.rule_group_reference_statement.excluded_rule` attributes.

The `statement.rule_group_reference_statement.rule_action_override` attribute has been added.

## resource/aws_wafv2_web_acl_logging_configuration

We removed the `redacted_fields.all_query_arguments`, `redacted_fields.body` and `redacted_fields.single_query_argument` attributes.

## data-source/aws_api_gateway_rest_api

The `minimum_compression_size` attribute is now a String type, allowing it to be computed when set via the `body` attribute.

## data-source/aws_db_instance

We removed the `db_security_groups` attribute as part of the EC2 Classic retirement.

## data-source/aws_connect_hours_of_operation

We removed the `hours_of_operation_arn` attribute.

## data-source/aws_elasticache_cluster

We removed the `security_group_names` attribute as part of the EC2 Classic retirement.

## data-source/aws_identitystore_group

We removed the `filter` argument.

## data-source/aws_identitystore_user

We removed the `filter` argument.

## data-source/aws_launch_configuration

We removed the `vpc_classic_link_id` and `vpc_classic_link_security_groups` attribute as part of the EC2 Classic retirement.

## data-source/aws_opensearch_domain

The `kibana_endpoint` attribute has been deprecated. All configurations using `kibana_endpoint` should be updated to use the `dashboard_endpoint` attribute instead.

## data-source/aws_quicksight_data_set

The `tags_all` attribute has been deprecated and will be removed in a future version.

## data-source/aws_redshift_cluster

We removed the `cluster_security_groups` attribute as part of the EC2 Classic retirement.

## data-source/aws_redshift_service_account

[AWS document](https://docs.aws.amazon.com/redshift/latest/mgmt/db-auditing.html#db-auditing-bucket-permissions) that [a service principal name](https://docs.aws.amazon.com/IAM/latest/UserGuide/reference_policies_elements_principal.html#principal-services) be used instead of AWS account ID in any relevant IAM policy.
The [`aws_redshift_service_account`](/docs/providers/aws/d/redshift_service_account.html) data source should now be considered deprecated and will be removed in a future version.

## data-source/aws_service_discovery_service

The `tags_all` attribute has been deprecated and will be removed in a future version.

## data-source/aws_secretsmanager_secret

We removed the `rotation_enabled`, `rotation_lambda_arn` and `rotation_rules` attributes.

## data-source/aws_subnet_ids

We removed the `aws_subnet_ids` data source. Use the [`aws_subnets`](/docs/providers/aws/d/subnets.html) data source instead.

