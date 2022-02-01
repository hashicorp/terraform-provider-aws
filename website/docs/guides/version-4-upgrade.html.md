---
subcategory: ""
layout: "aws"
page_title: "Terraform AWS Provider Version 4 Upgrade Guide"
description: |-
  Terraform AWS Provider Version 4 Upgrade Guide
---

# Terraform AWS Provider Version 4 Upgrade Guide

Version 4.0.0 of the AWS provider for Terraform is a major release and includes some changes that you will need to consider when upgrading. This guide is intended to help with that process and focuses only on changes from version 3.X to version 4.0.0. See the [Version 3 Upgrade Guide](/docs/providers/aws/guides/version-3-upgrade.html) for information about upgrading from 1.X to version 3.0.0.

Most of the changes outlined in this guide have been previously marked as deprecated in the Terraform plan/apply output throughout previous provider releases. These changes, such as deprecation notices, can always be found in the [Terraform AWS Provider CHANGELOG](https://github.com/hashicorp/terraform-provider-aws/blob/main/CHANGELOG.md).

~> **NOTE:** Version 4.0.0 of the AWS Provider will be the last major version to support [EC2-Classic resources](#ec2-classic-resource-and-data-source-support) as AWS plans to fully retire EC2-Classic Networking. See the [AWS News Blog](https://aws.amazon.com/blogs/aws/ec2-classic-is-retiring-heres-how-to-prepare/) for additional details.

~> **NOTE:** Version 4.0.0 and 4.X versions of the AWS Provider will be the last versions compatible with Terraform 0.12-0.14.

Upgrade topics:

<!-- TOC depthFrom:2 depthTo:2 -->

- [Provider Version Configuration](#provider-version-configuration)
- [Full Resource Lifecycle of Default Resources](#full-resource-lifecycle-of-default-resources)
    - [Resource: aws_default_subnet](#resource-aws_default_subnet)
    - [Resource: aws_default_vpc](#resource-aws_default_vpc)
- [Plural Data Source Behavior](#plural-data-source-behavior)
- [Data Source: aws_cloudwatch_log_group](#data-source-aws_cloudwatch_log_group)
- [Data Source: aws_subnet_ids](#data-source-aws_subnet_ids)
- [Resource: aws_batch_compute_environment](#resource-aws_batch_compute_environment)
- [Resource: aws_elasticache_cluster](#resource-aws_elasticache_cluster)
- [Resource: aws_elasticache_replication_group](#resource-aws_elasticache_replication_group)
- [Resource: aws_network_interface](#resource-aws_network_interface)
- [Resource: aws_s3_bucket](#resource-aws_s3_bucket)
- [Resource: aws_vpn_connection](#resource-aws_vpn_connection)

<!-- /TOC -->

Additional Topics:

<!-- TOC depthFrom:2 depthTo:2 -->

- [EC2-Classic resource and data source support](#ec2-classic-resource-and-data-source-support)

<!-- /TOC -->


## Provider Version Configuration

!> **WARNING:** This topic is placeholder documentation until version 4.0.0 is released.

-> Before upgrading to version 4.0.0, it is recommended to upgrade to the most recent 3.X version of the provider and ensure that your environment successfully runs [`terraform plan`](https://www.terraform.io/docs/commands/plan.html) without unexpected changes or deprecation notices.

It is recommended to use [version constraints when configuring Terraform providers](https://www.terraform.io/docs/configuration/providers.html#provider-versions). If you are following that recommendation, update the version constraints in your Terraform configuration and run [`terraform init`](https://www.terraform.io/docs/commands/init.html) to download the new version.

For example, given this previous configuration:

```terraform
provider "aws" {
  # ... other configuration ...

  version = "~> 3.74"
}
```

Update to latest 4.X version:

```terraform
provider "aws" {
  # ... other configuration ...

  version = "~> 4.0"
}
```

## Plural Data Source Behavior

The following plural data sources are now consistent with [Provider Design](https://github.com/hashicorp/terraform-provider-aws/blob/main/docs/contributing/provider-design.md#data-sources)
such that they no longer return an error if zero results are found.

* [aws_cognito_user_pools](/docs/providers/aws/d/cognito_user_pools.html)
* [aws_db_event_categories](/docs/providers/aws/d/db_event_categories.html)
* [aws_ebs_volumes](/docs/providers/aws/d/ebs_volumes.html)
* [aws_ec2_coip_pools](/docs/providers/aws/d/ec2_coip_pools.html)
* [aws_ec2_local_gateway_route_tables](/docs/providers/aws/d/ec2_local_gateway_route_tables.html)
* [aws_ec2_local_gateway_virtual_interface_groups](/docs/providers/aws/d/ec2_local_gateway_virtual_interface_groups.html)
* [aws_ec2_local_gateways](/docs/providers/aws/d/ec2_local_gateways.html)
* [aws_ec2_transit_gateway_route_tables](/docs/providers/aws/d/ec2_transit_gateway_route_tables.html)
* [aws_efs_access_points](/docs/providers/aws/d/efs_access_points.html)
* [aws_emr_release_labels](/docs/providers/aws/d/emr_release_labels.markdown)
* [aws_inspector_rules_packages](/docs/providers/aws/d/inspector_rules_packages.html)
* [aws_ip_ranges](/docs/providers/aws/d/ip_ranges.html)
* [aws_network_acls](/docs/providers/aws/d/network_acls.html)
* [aws_route_tables](/docs/providers/aws/d/route_tables.html)
* [aws_security_groups](/docs/providers/aws/d/security_groups.html)
* [aws_ssoadmin_instances](/docs/providers/aws/d/ssoadmin_instances.html)
* [aws_vpcs](/docs/providers/aws/d/vpcs.html)
* [aws_vpc_peering_connections](/docs/providers/aws/d/vpc_peering_connections.html)

## Full Resource Lifecycle of Default Resources

Default subnets and vpcs can now do full resource lifecycle operations such that resource
creation and deletion are now supported.

### Resource: aws_default_subnet

### Resource: aws_default_vpc

## Data Source: aws_cloudwatch_log_group

## Data Source: aws_subnet_ids

The `aws_subnet_ids` data source has been deprecated and will be removed removed in a future version. Use the `aws_subnets` data source instead.

For example, change a configuration such as

```hcl
data "aws_subnet_ids" "example" {
  vpc_id = var.vpc_id
}

data "aws_subnet" "example" {
  for_each = data.aws_subnet_ids.example.ids
  id       = each.value
}

output "subnet_cidr_blocks" {
  value = [for s in data.aws_subnet.example : s.cidr_block]
}
```

to

```hcl
data "aws_subnets" "example" {
  filter {
    name   = "vpc-id"
    values = [var.vpc_id]
  }
}

data "aws_subnet" "example" {
  for_each = data.aws_subnets.example.ids
  id       = each.value
}

output "subnet_cidr_blocks" {
  value = [for s in data.aws_subnet.example : s.cidr_block]
}
```

## Resource: aws_batch_compute_environment

No `compute_resources` can be specified when `type` is `UNMANAGED`.

Previously a configuration such as

```hcl
resource "aws_batch_compute_environment" "test" {
  compute_environment_name = "test"

  compute_resources {
    instance_role = aws_iam_instance_profile.ecs_instance.arn
    instance_type = [
      "c4.large",
    ]
    max_vcpus = 16
    min_vcpus = 0
    security_group_ids = [
      aws_security_group.test.id
    ]
    subnets = [
      aws_subnet.test.id
    ]
    type = "EC2"
  }

  service_role = aws_iam_role.batch_service.arn
  type         = "UNMANAGED"
}
```

could be applied and any compute resources were ignored.

Now this configuration is invalid and will result in an error during plan.

To resolve this error simply remove or comment out the `compute_resources` configuration block.

```hcl
resource "aws_batch_compute_environment" "test" {
  compute_environment_name = "test"

  service_role = aws_iam_role.batch_service.arn
  type         = "UNMANAGED"
}
```

## Resource: aws_elasticache_cluster

## Resource: aws_elasticache_replication_group

## Resource: aws_network_interface

## Resource: aws_s3_bucket

!> **WARNING:** This topic is placeholder documentation.

## Resource: aws_vpn_connection

Default values

## EC2-Classic Resource and Data Source Support

While an upgrade to this major version will not directly impact EC2-Classic resources configured with Terraform,
it is important to keep in the mind the following AWS Provider resources will eventually no longer
be compatible with EC2-Classic as AWS completes their EC2-Classic networking retirement (expected around August 15, 2022).

* Running or stopped [EC2 instances](/docs/providers/aws/r/instance.html)
* Running or stopped [RDS database instances](/docs/providers/aws/r/db_instance.html)
* [Elastic IP addresses](/docs/providers/aws/r/eip.html)
* [Classic Load Balancers](/docs/providers/aws/r/lb.html)
* [Redshift clusters](/docs/providers/aws/r/redshift_cluster.html)
* [Elastic Beanstalk environments](/docs/providers/aws/r/elastic_beanstalk_environment.html)
* [EMR clusters](/docs/providers/aws/r/emr_cluster.html)
* [AWS Data Pipelines pipelines](/docs/providers/aws/r/datapipeline_pipeline.html)
* [ElastiCache clusters](/docs/providers/aws/r/elasticache_cluster.html)
* [Spot Requests](/docs/providers/aws/r/spot_instance_request.html)
* [Capacity Reservations](/docs/providers/aws/r/ec2_capacity_reservation.html)