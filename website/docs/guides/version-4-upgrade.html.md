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

~> **NOTE:** Version 4.0.0 and 4.x.x versions of the AWS Provider will be the last versions compatible with Terraform 0.12-0.15.

Upgrade topics:

<!-- TOC depthFrom:2 depthTo:2 -->

- [Provider Version Configuration](#provider-version-configuration)
- [New Provider Arguments](#new-provider-arguments)
- [Full Resource Lifecycle of Default Resources](#full-resource-lifecycle-of-default-resources)
    - [Resource: aws_default_subnet](#resource-aws_default_subnet)
    - [Resource: aws_default_vpc](#resource-aws_default_vpc)
- [Plural Data Source Behavior](#plural-data-source-behavior)
- [Empty Strings Not Valid For Certain Resources](#empty-strings-not-valid-for-certain-resources)
    - [Resource: aws_cloudwatch_event_target (Empty String)](#resource-aws_cloudwatch_event_target-empty-string)
    - [Resource: aws_customer_gateway](#resource-aws_customer_gateway)
    - [Resource: aws_default_network_acl](#resource-aws_default_network_acl)
    - [Resource: aws_default_route_table](#resource-aws_default_route_table)
    - [Resource: aws_default_vpc (Empty String)](#resource-aws_default_vpc-empty-string)
    - [Resource: aws_efs_mount_target](#resource-aws_efs_mount_target)
    - [Resource: aws_elasticsearch_domain](#resource-aws_elasticsearch_domain)
    - [Resource: aws_instance](#resource-aws_instance)
    - [Resource: aws_network_acl](#resource-aws_network_acl)
    - [Resource: aws_route](#resource-aws_route)
    - [Resource: aws_route_table](#resource-aws_route_table)
    - [Resource: aws_vpc](#resource-aws_vpc)
    - [Resource: aws_vpc_ipv6_cidr_block_association](#resource-aws_vpc_ipv6_cidr_block_association)
- [Data Source: aws_cloudwatch_log_group](#data-source-aws_cloudwatch_log_group)
- [Data Source: aws_subnet_ids](#data-source-aws_subnet_ids)
- [Data Source: aws_s3_bucket_object](#data-source-aws_s3_bucket_object)
- [Data Source: aws_s3_bucket_objects](#data-source-aws_s3_bucket_objects)
- [Resource: aws_batch_compute_environment](#resource-aws_batch_compute_environment)
- [Resource: aws_cloudwatch_event_target](#resource-aws_cloudwatch_event_target)
- [Resource: aws_elasticache_cluster](#resource-aws_elasticache_cluster)
- [Resource: aws_elasticache_global_replication_group](#resource-aws_elasticache_global_replication_group)
- [Resource: aws_elasticache_replication_group](#resource-aws_elasticache_replication_group)
- [Resource: aws_fsx_ontap_storage_virtual_machine](#resource-aws_fsx_ontap_storage_virtual_machine)
- [Resource: aws_lb_target_group](#resource-aws_lb_target_group)
- [Resource: aws_network_interface](#resource-aws_network_interface)
- [Resource: aws_s3_bucket](#resource-aws_s3_bucket)
- [Resource: aws_s3_bucket_object](#resource-aws_s3_bucket_object)
- [Resource: aws_spot_instance_request](#resource-aws_spot_instance_request)

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

## New Provider Arguments

Version 4.0.0 adds these new provider arguments:

* `ec2_metadata_service_endpoint` - Address of the EC2 metadata service (IMDS) endpoint to use. Can also be set with the `AWS_EC2_METADATA_SERVICE_ENDPOINT` environment variable.
* `ec2_metadata_service_endpoint_mode` - Mode to use in communicating with the metadata service. Valid values are `IPv4` and `IPv6`. Can also be set with the `AWS_EC2_METADATA_SERVICE_ENDPOINT_MODE` environment variable.
* `use_dualstack_endpoint` - Force the provider to resolve endpoints with DualStack capability. Can also be set with the `AWS_USE_DUALSTACK_ENDPOINT` environment variable or in a shared config file (`use_dualstack_endpoint`).
* `use_fips_endpoint` - Force the provider to resolve endpoints with FIPS capability. Can also be set with the `AWS_USE_FIPS_ENDPOINT` environment variable or in a shared config file (`use_fips_endpoint`).

~> **NOTE:** Using the `AWS_METADATA_URL` environment variable has been deprecated in Terraform AWS Provider v4.0.0 and support will be removed in a future version. Change any scripts or environments using `AWS_METADATA_URL` to instead use `AWS_EC2_METADATA_SERVICE_ENDPOINT`.

For example, in previous versions, to use FIPS endpoints, you would need to provide all the FIPS endpoints that you wanted to use in the `endpoints` configuration block:

```terraform
provider "aws" {
  endpoints {
    ec2 = "https://ec2-fips.us-west-2.amazonaws.com"
    s3  = "https://s3-fips.us-west-2.amazonaws.com"
    sts = "https://sts-fips.us-west-2.amazonaws.com"
  }
}
```

In v4.0.0, you can still set endpoints in the same way. However, you can instead use the `use_fips_endpoint` argument to have the provider automatically resolve FIPS endpoints for all supported services:

```terraform
provider "aws" {
  use_fips_endpoint = true
}
```

Note that the provider can only resolve FIPS endpoints where AWS provides FIPS support. Support depends on the service and may include `us-east-1`, `us-east-2`, `us-west-1`, `us-west-2`, `us-gov-east-1`, `us-gov-west-1`, and `ca-central-1`. For more information, see [Federal Information Processing Standard (FIPS) 140-2](https://aws.amazon.com/compliance/fips/).

## Full Resource Lifecycle of Default Resources

Default subnets and vpcs can now do full resource lifecycle operations such that resource
creation and deletion are now supported.

### Resource: aws_default_subnet

The `aws_default_subnet` resource behaves differently from normal resources in that if a default subnet exists in the specified Availability Zone, Terraform does not _create_ this resource, but instead "adopts" it into management.
If no default subnet exists, Terraform creates a new default subnet.
By default, `terraform destroy` does not delete the default subnet but does remove the resource from Terraform state.
Set the `force_destroy` argument to `true` to delete the default subnet.

For example, given this previous configuration with no existing default subnet:

```terraform
terraform {
  required_providers {
    aws = {
      source = "hashicorp/aws"
    }
  }
  required_version = ">= 0.13"
}

provider "aws" {
  region = "eu-west-2"
}

resource "aws_default_subnet" "default" {}
```

The following error was thrown on `terraform apply`:

```
│ Error: Default subnet not found.
│
│   with aws_default_subnet.default,
│   on main.tf line 5, in resource "aws_default_subnet" "default":
│    5: resource "aws_default_subnet" "default" {}
```

Now after upgrading, the above configuration will apply successfully.

To delete the default subnet, the above configuration should be updated to:

```terraform
terraform {
  required_providers {
    aws = {
      source = "hashicorp/aws"
    }
  }
  required_version = ">= 0.13"
}

resource "aws_default_subnet" "default" {
  force_destroy = true
}
```

### Resource: aws_default_vpc

The `aws_default_vpc` resource behaves differently from normal resources in that if a default VPC exists, Terraform does not _create_ this resource, but instead "adopts" it into management.
If no default VPC exists, Terraform creates a new default VPC, which leads to the implicit creation of [other resources](https://docs.aws.amazon.com/vpc/latest/userguide/default-vpc.html#default-vpc-components).
By default, `terraform destroy` does not delete the default VPC but does remove the resource from Terraform state.
Set the `force_destroy` argument to `true` to delete the default VPC.

For example, given this previous configuration with no existing default VPC:

```terraform
terraform {
  required_providers {
    aws = {
      source = "hashicorp/aws"
    }
  }
  required_version = ">= 0.13"
}

resource "aws_default_vpc" "default" {}
```

The following error was thrown on `terraform apply`:

```
│ Error: No default VPC found in this region.
│
│   with aws_default_vpc.default,
│   on main.tf line 5, in resource "aws_default_vpc" "default":
│    5: resource "aws_default_vpc" "default" {}
```

Now after upgrading, the above configuration will apply successfully.

To delete the default VPC, the above configuration should be updated to:

```terraform
terraform {
  required_providers {
    aws = {
      source = "hashicorp/aws"
    }
  }
  required_version = ">= 0.13"
}

resource "aws_default_vpc" "default" {
  force_destroy = true
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

## Empty Strings Not Valid For Certain Resources

First, this is a breaking change but should affect very few configurations.

Second, the motivation behind this change is that previously, you might set an argument to `""` to explicitly convey it is empty. However, with the introduction of `null` in Terraform 0.12 and to prepare for continuing enhancements that distinguish between unset arguments and those that have a value, including an empty string (`""`), we are moving away from this use of zero values. We ask practitioners to either use `null` instead or remove the arguments that are set to `""`.

### Resource: aws_cloudwatch_event_target (Empty String)

Previously, `ecs_target.0.launch_type` could be set to `""`. However, the value `""` is no longer valid. Now, set the argument to `null` (_e.g._, `launch_type = null`) or remove the empty-string configuration.

For example, this type of configuration is now not valid:

```terraform
resource "aws_cloudwatch_event_target" "example" {
  # ...
  ecs_target {
    task_count          = 1
    task_definition_arn = aws_ecs_task_definition.task.arn
    launch_type         = ""
    # ...
  }
}
```

In this updated and valid configuration, we set `launch_type` to `null`:

```terraform
resource "aws_cloudwatch_event_target" "example" {
  # ...
  ecs_target {
    task_count          = 1
    task_definition_arn = aws_ecs_task_definition.task.arn
    launch_type         = null
    # ...
  }
}
```

### Resource: aws_customer_gateway

Previously, `ip_address` could be set to `""`, which would result in an AWS error. However, this value is no longer accepted by the provider.

### Resource: aws_default_network_acl

Previously, `egress.*.cidr_block`, `egress.*.ipv6_cidr_block`, `ingress.*.cidr_block`, and `ingress.*.ipv6_cidr_block` could be set to `""`. However, the value `""` is no longer valid. Now, set the argument to `null` (_e.g._, `ipv6_cidr_block = null`) or remove the empty-string configuration.

For example, this type of configuration is now not valid:

```terraform
resource "aws_default_network_acl" "example" {
  # ...
  egress {
    cidr_block      = "0.0.0.0/0"
    ipv6_cidr_block = ""
    # ...
  }
}
```

In this updated and valid configuration, we remove the empty-string configuration:

```terraform
resource "aws_default_network_acl" "example" {
  # ...
  egress {
    cidr_block = "0.0.0.0/0"
    # ...
  }
}
```

### Resource: aws_default_route_table

Previously, `route.*.cidr_block` and `route.*.ipv6_cidr_block` could be set to `""`. However, the value `""` is no longer valid. Now, set the argument to `null` (_e.g._, `ipv6_cidr_block = null`) or remove the empty-string configuration.

For example, this type of configuration is now not valid:

```terraform
resource "aws_default_route_table" "example" {
  # ...
  route {
    cidr_block      = local.ipv6 ? "" : local.destination
    ipv6_cidr_block = local.ipv6 ? local.destination_ipv6 : ""
  }
}
```

In this updated and valid configuration, we use `null` instead of an empty string (`""`):

```terraform
resource "aws_default_route_table" "example" {
  # ...
  route {
    cidr_block      = local.ipv6 ? null : local.destination
    ipv6_cidr_block = local.ipv6 ? local.destination_ipv6 : null
  }
}
```

### Resource: aws_default_vpc (Empty String)

Previously, `ipv6_cidr_block` could be set to `""`. However, the value `""` is no longer valid. Now, set the argument to `null` (_e.g._, `ipv6_cidr_block = null`) or remove the empty-string configuration.

### Resource: aws_instance

Previously, `private_ip` could be set to `""`. However, the value `""` is no longer valid. Now, set the argument to `null` (_e.g._, `private_ip = null`) or remove the empty-string configuration.

For example, this type of configuration is now not valid:

```terraform
resource "aws_instance" "example" {
  instance_type = "t2.micro"
  private_ip    = ""
}
```

In this updated and valid configuration, we remove the empty-string configuration:

```terraform
resource "aws_instance" "example" {
  instance_type = "t2.micro"
}
```

### Resource: aws_efs_mount_target

Previously, `ip_address` could be set to `""`. However, the value `""` is no longer valid. Now, set the argument to `null` (_e.g._, `ip_address = null`) or remove the empty-string configuration.

For example, this type of configuration is now not valid: `ip_address = ""`.

### Resource: aws_elasticsearch_domain

Previously, `ebs_options.0.volume_type` could be set to `""`. However, the value `""` is no longer valid. Now, set the argument to `null` (_e.g._, `volume_type = null`) or remove the empty-string configuration.

For example, this type of configuration is now not valid:

```terraform
resource "aws_elasticsearch_domain" "example" {
  # ...
  ebs_options {
    ebs_enabled = true
    volume_size = var.volume_size
    volume_type = var.volume_size > 0 ? local.volume_type : ""
  }
}
```

In this updated and valid configuration, we use `null` instead of `""`:

```terraform
resource "aws_elasticsearch_domain" "example" {
  # ...
  ebs_options {
    ebs_enabled = true
    volume_size = var.volume_size
    volume_type = var.volume_size > 0 ? local.volume_type : null
  }
}
```

### Resource: aws_network_acl

Previously, `egress.*.cidr_block`, `egress.*.ipv6_cidr_block`, `ingress.*.cidr_block`, and `ingress.*.ipv6_cidr_block` could be set to `""`. However, the value `""` is no longer valid. Now, set the argument to `null` (_e.g._, `ipv6_cidr_block = null`) or remove the empty-string configuration.

For example, this type of configuration is now not valid:

```terraform
resource "aws_network_acl" "example" {
  # ...
  egress {
    cidr_block      = "0.0.0.0/0"
    ipv6_cidr_block = ""
    # ...
  }
}
```

In this updated and valid configuration, we remove the empty-string configuration:

```terraform
resource "aws_network_acl" "example" {
  # ...
  egress {
    cidr_block = "0.0.0.0/0"
    # ...
  }
}
```

### Resource: aws_route

Previously, `destination_cidr_block` and `destination_ipv6_cidr_block` could be set to `""`. However, the value `""` is no longer valid. Now, set the argument to `null` (_e.g._, `destination_ipv6_cidr_block = null`) or remove the empty-string configuration.

In addition, now exactly one of `destination_cidr_block`, `destination_ipv6_cidr_block`, and `destination_prefix_list_id` can be set.

For example, this type of configuration for `aws_route` is now not valid:

```terraform
resource "aws_route" "example" {
  route_table_id = aws_route_table.example.id
  gateway_id     = aws_internet_gateway.example.id

  destination_cidr_block      = local.ipv6 ? "" : local.destination
  destination_ipv6_cidr_block = local.ipv6 ? local.destination_ipv6 : ""
}
```

In this updated and valid configuration, we use `null` instead of an empty-string (`""`):

```terraform
resource "aws_route" "example" {
  route_table_id = aws_route_table.example.id
  gateway_id     = aws_internet_gateway.example.id

  destination_cidr_block      = local.ipv6 ? null : local.destination
  destination_ipv6_cidr_block = local.ipv6 ? local.destination_ipv6 : null
}
```

### Resource: aws_route_table

Previously, `route.*.cidr_block` and `route.*.ipv6_cidr_block` could be set to `""`. However, the value `""` is no longer valid. Now, set the argument to `null` (_e.g._, `ipv6_cidr_block = null`) or remove the empty-string configuration.

For example, this type of configuration is now not valid:

```terraform
resource "aws_route_table" "example" {
  # ...
  route {
    cidr_block      = local.ipv6 ? "" : local.destination
    ipv6_cidr_block = local.ipv6 ? local.destination_ipv6 : ""
  }
}
```

In this updated and valid configuration, we used `null` instead of an empty-string (`""`):

```terraform
resource "aws_route_table" "example" {
  # ...
  route {
    cidr_block      = local.ipv6 ? null : local.destination
    ipv6_cidr_block = local.ipv6 ? local.destination_ipv6 : null
  }
}
```

### Resource: aws_vpc

Previously, `ipv6_cidr_block` could be set to `""`. However, the value `""` is no longer valid. Now, set the argument to `null` (_e.g._, `ipv6_cidr_block = null`) or remove the empty-string configuration.

For example, this type of configuration is now not valid:

```terraform
resource "aws_vpc" "example" {
  cidr_block      = "10.1.0.0/16"
  ipv6_cidr_block = ""
}
```

In this updated and valid configuration, we remove `ipv6_cidr_block`:

```terraform
resource "aws_vpc" "example" {
  cidr_block      = "10.1.0.0/16"
}
```

### Resource: aws_vpc_ipv6_cidr_block_association

Previously, `ipv6_cidr_block` could be set to `""`. However, the value `""` is no longer valid. Now, set the argument to `null` (_e.g._, `ipv6_cidr_block = null`) or remove the empty-string configuration.

## Data Source: aws_cloudwatch_log_group

### Removal of arn Wildcard Suffix

Previously, the data source returned the Amazon Resource Name (ARN) directly from the API, which included a `:*` suffix to denote all CloudWatch Log Streams under the CloudWatch Log Group. Most other AWS resources that return ARNs and many other AWS services do not use the `:*` suffix. The suffix is now automatically removed. For example, the data source previously returned an ARN such as `arn:aws:logs:us-east-1:123456789012:log-group:/example:*` but will now return `arn:aws:logs:us-east-1:123456789012:log-group:/example`.

Workarounds, such as using `replace()` as shown below, should be removed:

```terraform
data "aws_cloudwatch_log_group" "example" {
  name = "example"
}
resource "aws_datasync_task" "example" {
  # ... other configuration ...
  cloudwatch_log_group_arn = replace(data.aws_cloudwatch_log_group.example.arn, ":*", "")
}
```

Removing the `:*` suffix is a breaking change for some configurations. Fix these configurations using string interpolations as demonstrated below. For example, this configuration is now broken:

```terraform
data "aws_iam_policy_document" "ad-log-policy" {
  statement {
    actions = [
      "logs:CreateLogStream",
      "logs:PutLogEvents"
    ]
    principals {
      identifiers = ["ds.amazonaws.com"]
      type        = "Service"
    }
    resources = [data.aws_cloudwatch_log_group.example.arn]
    effect = "Allow"
  }
}
```

An updated configuration:

```terraform
data "aws_iam_policy_document" "ad-log-policy" {
  statement {
    actions = [
      "logs:CreateLogStream",
      "logs:PutLogEvents"
    ]
    principals {
      identifiers = ["ds.amazonaws.com"]
      type        = "Service"
    }
    resources = ["${data.aws_cloudwatch_log_group.example.arn}:*"]
    effect = "Allow"
  }
}
```

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

## Data Source: aws_s3_bucket_object

The `aws_s3_bucket_object` data source is deprecated and will be removed in a future version. Use `aws_s3_object` instead, where new features and fixes will be added.

## Data Source: aws_s3_bucket_objects

The `aws_s3_bucket_objects` data source is deprecated and will be removed in a future version. Use `aws_s3_objects` instead, where new features and fixes will be added.

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

## Resource: aws_cloudwatch_event_target

### Removal of `ecs_target` `launch_type` default value

Previously, the `ecs_target` `launch_type` argument defaulted to `EC2` if no value was configured in Terraform.

Workarounds, such as using the empty string `""` as shown below, should be removed:

```terraform
resource "aws_cloudwatch_event_target" "test" {
  arn      = aws_ecs_cluster.test.id
  rule     = aws_cloudwatch_event_rule.test.id
  role_arn = aws_iam_role.test.arn
  ecs_target {
    launch_type         = ""
    task_count          = 1
    task_definition_arn = aws_ecs_task_definition.task.arn
    network_configuration {
      subnets = [aws_subnet.subnet.id]
    }
  }
}
```

An updated configuration:

```terraform
resource "aws_cloudwatch_event_target" "test" {
  arn      = aws_ecs_cluster.test.id
  rule     = aws_cloudwatch_event_rule.test.id
  role_arn = aws_iam_role.test.arn
  ecs_target {
    task_count          = 1
    task_definition_arn = aws_ecs_task_definition.task.arn
    network_configuration {
      subnets = [aws_subnet.subnet.id]
    }
  }
}
```

## Resource: aws_elasticache_cluster

### Error raised if neither `engine` nor `replication_group_id` is specified

Previously, when neither `engine` nor `replication_group_id` was specified, Terraform would not prevent users from applying the invalid configuration.
Now, this will produce an error similar to the below:

```
Error: Invalid combination of arguments

          with aws_elasticache_cluster.example,
          on terraform_plugin_test.tf line 2, in resource "aws_elasticache_cluster" "example":
           2: resource "aws_elasticache_cluster" "example" {

        "replication_group_id": one of `engine,replication_group_id` must be
        specified

        Error: Invalid combination of arguments

          with aws_elasticache_cluster.example,
          on terraform_plugin_test.tf line 2, in resource "aws_elasticache_cluster" "example":
           2: resource "aws_elasticache_cluster" "example" {

        "engine": one of `engine,replication_group_id` must be specified
```

Configuration that depend on the previous behavior will need to be updated.

## Resource: aws_elasticache_global_replication_group

### actual_engine_version Attribute removal

Switch your Terraform configuration to the `engine_version_actual` attribute instead.

For example, given this previous configuration:

```terraform
output "elasticache_global_replication_group_version_result" {
  value = aws_elasticache_global_replication_group.example.actual_engine_version
}
```

An updated configuration:

```terraform
output "elasticache_global_replication_group_version_result" {
  value = aws_elasticache_global_replication_group.example.engine_version_actual
}
```

## Resource: aws_elasticache_replication_group

!> **WARNING:** This topic is placeholder documentation.

## Resource: aws_fsx_ontap_storage_virtual_machine

We removed the misspelled argument `active_directory_configuration.0.self_managed_active_directory_configuration.0.organizational_unit_distinguidshed_name` that was previously deprecated. Use `active_directory_configuration.0.self_managed_active_directory_configuration.0.organizational_unit_distinguished_name` now instead. Terraform will automatically migrate the state to `active_directory_configuration.0.self_managed_active_directory_configuration.0.organizational_unit_distinguished_name` during planning.

## Resource: aws_lb_target_group


For `protocol = "TCP"`, `stickiness.type` can no longer be set to `lb_cookie` even when `enabled = false`. Instead, either change the `protocol` to `"HTTP"` or `"HTTPS"`, or change `stickiness.type` to `"source_ip"`.

For example, this configuration is no longer valid:

```terraform
resource "aws_lb_target_group" "test" {
  port        = 25
  protocol    = "TCP"
  vpc_id      = aws_vpc.test.id

  stickiness {
    type    = "lb_cookie"
    enabled = false
  }
}
```

To fix this, we change the `stickiness.type` to `"source_ip"`.

```terraform
resource "aws_lb_target_group" "test" {
  port        = 25
  protocol    = "TCP"
  vpc_id      = aws_vpc.test.id

  stickiness {
    type    = "source_ip"
    enabled = false
  }
}
```


## Resource: aws_network_interface

!> **WARNING:** This topic is placeholder documentation.

## Resource: aws_s3_bucket

To help distribute the management of S3 bucket settings via independent resources, various arguments and attributes in the `aws_s3_bucket` resource
have become read-only. Configurations dependent on these arguments should be updated to use the corresponding `aws_s3_bucket_*` resource.
Once updated, new `aws_s3_bucket_*` resources should be imported into Terraform state.

### `acceleration_status` Argument deprecation

Switch your Terraform configuration to the `aws_s3_bucket_accelerate_configuration` resource instead.

For example, given this previous configuration:

```terraform
resource "aws_s3_bucket" "example" {
  # ... other configuration ...
  acceleration_status = "Enabled"
}
```

It will receive the following error after upgrading:

```
│ Error: Value for unconfigurable attribute
│
│   with aws_s3_bucket.example,
│   on main.tf line 1, in resource "aws_s3_bucket" "example":
│    1: resource "aws_s3_bucket" "example" {
│
│ Can't configure a value for "acceleration_status": its value will be decided automatically based on the result of applying this configuration.
```

Since the `acceleration_status` argument changed to read-only, the recommendation is to update the configuration to use the `aws_s3_bucket_accelerate_configuration`
resource and remove any reference to `acceleration_status` in the `aws_s3_bucket` resource:

```terraform
resource "aws_s3_bucket" "example" {
  # ... other configuration ...
}

resource "aws_s3_bucket_accelerate_configuration" "example" {
  bucket = aws_s3_bucket.example.id
  status = "Enabled"
}
```

It is then recommended running `terraform import` on each new resource to prevent data loss, e.g.

```shell
$ terraform import aws_s3_bucket_accelerate_configuration.example example
aws_s3_bucket_accelerate_configuration.example: Importing from ID "example"...
aws_s3_bucket_accelerate_configuration.example: Import prepared!
  Prepared aws_s3_bucket_accelerate_configuration for import
aws_s3_bucket_accelerate_configuration.example: Refreshing state... [id=example]

Import successful!

The resources that were imported are shown above. These resources are now in
your Terraform state and will henceforth be managed by Terraform.
```

### `acl` Argument deprecation

Switch your Terraform configuration to the `aws_s3_bucket_acl` resource instead.

For example, given this previous configuration:

```terraform
resource "aws_s3_bucket" "example" {
  # ... other configuration ...
  acl = "private"
}
```

It will receive the following error after upgrading:

```
│ Error: Value for unconfigurable attribute
│
│   with aws_s3_bucket.example,
│   on main.tf line 1, in resource "aws_s3_bucket" "example":
│    1: resource "aws_s3_bucket" "example" {
│
│ Can't configure a value for "acl": its value will be decided automatically based on the result of applying this configuration.
```

Since the `acl` argument changed to read-only, the recommendation is to update the configuration to use the `aws_s3_bucket_acl`
resource and remove any reference to `acl` in the `aws_s3_bucket` resource:

```terraform
resource "aws_s3_bucket" "example" {
  # ... other configuration ...
}

resource "aws_s3_bucket_acl" "example" {
  bucket = aws_s3_bucket.example.id
  acl    = "private"
}
```

It is then recommended running `terraform import` on each new resource to prevent data loss, e.g.

```shell
$ terraform import aws_s3_bucket_acl.example example,private
aws_s3_bucket_acl.example: Importing from ID "example,private"...
aws_s3_bucket_acl.example: Import prepared!
  Prepared aws_s3_bucket_acl for import
aws_s3_bucket_acl.example: Refreshing state... [id=example,private]

Import successful!

The resources that were imported are shown above. These resources are now in
your Terraform state and will henceforth be managed by Terraform.
```

### `cors_rule` Argument deprecation

Switch your Terraform configuration to the `aws_s3_bucket_cors_configuration` resource instead.

For example, given this previous configuration:

```terraform
resource "aws_s3_bucket" "example" {
  # ... other configuration ...
  cors_rule {
    allowed_headers = ["*"]
    allowed_methods = ["PUT", "POST"]
    allowed_origins = ["https://s3-website-test.hashicorp.com"]
    expose_headers  = ["ETag"]
    max_age_seconds = 3000
  }
}
```

It will receive the following error after upgrading:

```
│ Error: Value for unconfigurable attribute
│
│   with aws_s3_bucket.example,
│   on main.tf line 1, in resource "aws_s3_bucket" "example":
│    1: resource "aws_s3_bucket" "example" {
│
│ Can't configure a value for "cors_rule": its value will be decided automatically based on the result of applying this configuration.
```

Since the `cors_rule` argument changed to read-only, the recommendation is to update the configuration to use the `aws_s3_bucket_cors_configuration`
resource and remove any references to `cors_rule` and its nested arguments in the `aws_s3_bucket` resource:

```terraform
resource "aws_s3_bucket" "example" {
  # ... other configuration ...
}

resource "aws_s3_bucket_cors_configuration" "example" {
  bucket = aws_s3_bucket.example.id

  cors_rule {
    allowed_headers = ["*"]
    allowed_methods = ["PUT", "POST"]
    allowed_origins = ["https://s3-website-test.hashicorp.com"]
    expose_headers  = ["ETag"]
    max_age_seconds = 3000
  }
}
```

It is then recommended running `terraform import` on each new resource to prevent data loss, e.g.

```shell
$ terraform import aws_s3_bucket_cors_configuration.example example
aws_s3_bucket_cors_configuration.example: Importing from ID "example"...
aws_s3_bucket_cors_configuration.example: Import prepared!
  Prepared aws_s3_bucket_cors_configuration for import
aws_s3_bucket_cors_configuration.example: Refreshing state... [id=example]

Import successful!

The resources that were imported are shown above. These resources are now in
your Terraform state and will henceforth be managed by Terraform.
```

### `grant` Argument deprecation

Switch your Terraform configuration to the `aws_s3_bucket_acl` resource instead.

For example, given this previous configuration:

```terraform
resource "aws_s3_bucket" "example" {
  # ... other configuration ...
  grant {
    id          = data.aws_canonical_user_id.current_user.id
    type        = "CanonicalUser"
    permissions = ["FULL_CONTROL"]
  }
  grant {
    type        = "Group"
    permissions = ["READ_ACP", "WRITE"]
    uri         = "http://acs.amazonaws.com/groups/s3/LogDelivery"
  }
}
```

It will receive the following error after upgrading:

```
│ Error: Value for unconfigurable attribute
│
│   with aws_s3_bucket.example,
│   on main.tf line 1, in resource "aws_s3_bucket" "example":
│    1: resource "aws_s3_bucket" "example" {
│
│ Can't configure a value for "grant": its value will be decided automatically based on the result of applying this configuration.
```

Since the `grant` argument changed to read-only, the recommendation is to update the configuration to use the `aws_s3_bucket_acl`
resource and remove any reference to `grant` in the `aws_s3_bucket` resource:

```terraform
resource "aws_s3_bucket" "example" {
  # ... other configuration ...
}

resource "aws_s3_bucket_acl" "example" {
  bucket = aws_s3_bucket.example.id
  
  access_control_policy {
    grant {
      grantee {
        id   = data.aws_canonical_user_id.current_user.id
        type = "CanonicalUser"
      }
      permission = "FULL_CONTROL"
    }

    grant {
      grantee {
        type = "Group"
        uri  = "http://acs.amazonaws.com/groups/s3/LogDelivery"
      }
      permission = "READ_ACP"
    }

    grant {
      grantee {
        type = "Group"
        uri  = "http://acs.amazonaws.com/groups/s3/LogDelivery"
      }
      permission = "WRITE"
    }

    owner {
      id = data.aws_canonical_user_id.current_user.id
    }
  }
}
```

It is then recommended running `terraform import` on each new resource to prevent data loss, e.g.

```shell
$ terraform import aws_s3_bucket_acl.example example
aws_s3_bucket_acl.example: Importing from ID "example"...
aws_s3_bucket_acl.example: Import prepared!
  Prepared aws_s3_bucket_acl for import
aws_s3_bucket_acl.example: Refreshing state... [id=example]

Import successful!

The resources that were imported are shown above. These resources are now in
your Terraform state and will henceforth be managed by Terraform.
```

### `lifecycle_rule` Argument deprecation

Switch your Terraform configuration to the `aws_s3_bucket_lifecycle_configuration` resource instead.

For example, given this previous configuration:

```terraform
resource "aws_s3_bucket" "example" {
  # ... other configuration ...
  lifecycle_rule {
    id      = "log"
    enabled = true
    prefix = "log/"
    tags = {
      rule      = "log"
      autoclean = "true"
    }
    transition {
      days          = 30
      storage_class = "STANDARD_IA"
    }
    transition {
      days          = 60
      storage_class = "GLACIER"
    }
    expiration {
      days = 90
    }
  }

  lifecycle_rule {
    id      = "tmp"
    prefix  = "tmp/"
    enabled = true
    expiration {
      date = "2022-12-31"
    }
  }
}
```

It will receive the following error after upgrading:

```
│ Error: Value for unconfigurable attribute
│
│   with aws_s3_bucket.example,
│   on main.tf line 1, in resource "aws_s3_bucket" "example":
│    1: resource "aws_s3_bucket" "example" {
│
│ Can't configure a value for "lifecycle_rule": its value will be decided automatically based on the result of applying this configuration.
```

Since the `lifecycle_rule` argument changed to read-only, the recommendation is to update the configuration to use the `aws_s3_bucket_lifecycle_configuration`
resource and remove any references to `lifecycle_rule` and its nested arguments in the `aws_s3_bucket` resource:

```terraform
resource "aws_s3_bucket" "example" {
  # ... other configuration ...
}

resource "aws_s3_bucket_lifecycle_configuration" "example" {
  bucket = aws_s3_bucket.example.id

  rule {
    id     = "log"
    status = "Enabled"

    filter {
      and {
        prefix = "log/"
        tags = {
          rule      = "log"
          autoclean = "true"
        }
      }
    }

    transition {
      days          = 30
      storage_class = "STANDARD_IA"
    }

    transition {
      days          = 60
      storage_class = "GLACIER"
    }

    expiration {
      days = 90
    }
  }

  rule {
    id = "tmp"

    filter {
      prefix  = "tmp/"
    }

    expiration {
      date = "2022-12-31T00:00:00Z"
    }

    status = "Enabled"
  }
}
```

It is then recommended running `terraform import` on each new resource to prevent data loss, e.g.

```shell
$ terraform import aws_s3_bucket_lifecycle_configuration.example example
aws_s3_bucket_lifecycle_configuration.example: Importing from ID "example"...
aws_s3_bucket_lifecycle_configuration.example: Import prepared!
  Prepared aws_s3_bucket_lifecycle_configuration for import
aws_s3_bucket_lifecycle_configuration.example: Refreshing state... [id=example]

Import successful!

The resources that were imported are shown above. These resources are now in
your Terraform state and will henceforth be managed by Terraform.
```

### `logging` Argument deprecation

Switch your Terraform configuration to the `aws_s3_bucket_logging` resource instead.

For example, given this previous configuration:

```terraform
resource "aws_s3_bucket" "log_bucket" {
  # ... other configuration ...
  bucket = "example-log-bucket"
}

resource "aws_s3_bucket" "example" {
  # ... other configuration ...
  logging {
    target_bucket = aws_s3_bucket.log_bucket.id
    target_prefix = "log/"
  }
}
```

It will receive the following error after upgrading:

```
│ Error: Value for unconfigurable attribute
│
│   with aws_s3_bucket.example,
│   on main.tf line 1, in resource "aws_s3_bucket" "example":
│    1: resource "aws_s3_bucket" "example" {
│
│ Can't configure a value for "logging": its value will be decided automatically based on the result of applying this configuration.
```

Since the `logging` argument changed to read-only, the recommendation is to update the configuration to use the `aws_s3_bucket_logging`
resource and remove any references to `logging` and its nested arguments in the `aws_s3_bucket` resource:

```terraform
resource "aws_s3_bucket" "log_bucket" {
  # ... other configuration ...
  bucket = "example-log-bucket"
}

resource "aws_s3_bucket" "example" {
  # ... other configuration ...
}

resource "aws_s3_bucket_logging" "example" {
  bucket        = aws_s3_bucket.example.id
  target_bucket = aws_s3_bucket.log_bucket.id
  target_prefix = "log/"
}
```

It is then recommended running `terraform import` on each new resource to prevent data loss, e.g.

```shell
$ terraform import aws_s3_bucket_logging.example example
aws_s3_bucket_logging.example: Importing from ID "example"...
aws_s3_bucket_logging.example: Import prepared!
  Prepared aws_s3_bucket_logging for import
aws_s3_bucket_logging.example: Refreshing state... [id=example]

Import successful!

The resources that were imported are shown above. These resources are now in
your Terraform state and will henceforth be managed by Terraform.
```

### `object_lock_configuration` `rule` Argument deprecation

Switch your Terraform configuration to the `aws_s3_bucket_object_lock_configuration` resource instead.

For example, given this previous configuration:

```terraform
resource "aws_s3_bucket" "example" {
  # ... other configuration ...
  object_lock_configuration {
    object_lock_enabled = "Enabled"

    rule {
      default_retention {
        mode = "COMPLIANCE"
        days = 3
      }
    }
  }
}
```

It will receive the following error after upgrading:

```
│ Error: Value for unconfigurable attribute
│
│   with aws_s3_bucket.example,
│   on main.tf line 1, in resource "aws_s3_bucket" "example":
│    1: resource "aws_s3_bucket" "example" {
│
│ Can't configure a value for "object_lock_configuration.0.rule": its value will be decided automatically based on the result of applying this configuration.
```

Since the `rule` argument of the `object_lock_configuration` configuration block changed to read-only, the recommendation is to update the configuration to use the `aws_s3_bucket_object_lock_configuration`
resource and remove any references to `rule` and its nested arguments in the `aws_s3_bucket` resource:

```terraform
resource "aws_s3_bucket" "example" {
  # ... other configuration ...
  object_lock_configuration {
    object_lock_enabled = "Enabled"
  }
}

resource "aws_s3_bucket_object_lock_configuration" "example" {
  bucket = aws_s3_bucket.example.id

  rule {
    default_retention {
      mode = "COMPLIANCE"
      days = 3
    }
  }
}
```

It is then recommended running `terraform import` on each new resource to prevent data loss, e.g.

```shell
$ terraform import aws_s3_bucket_object_lock_configuration.example example
aws_s3_bucket_object_lock_configuration.example: Importing from ID "example"...
aws_s3_bucket_object_lock_configuration.example: Import prepared!
  Prepared aws_s3_bucket_object_lock_configuration for import
aws_s3_bucket_object_lock_configuration.example: Refreshing state... [id=example]

Import successful!

The resources that were imported are shown above. These resources are now in
your Terraform state and will henceforth be managed by Terraform.
```

### `policy` Argument deprecation

Switch your Terraform configuration to the `aws_s3_bucket_policy` resource instead.

For example, given this previous configuration:

```terraform
resource "aws_s3_bucket" "accesslogs_bucket" {
  # ... other configuration ...
  policy = <<EOF
{
  "Id": "Policy1446577137248",
  "Statement": [
    {
      "Action": "s3:PutObject",
      "Effect": "Allow",
      "Principal": {
        "AWS": "${data.aws_elb_service_account.current.arn}"
      },
      "Resource": "arn:${data.aws_partition.current.partition}:s3:::example/*",
      "Sid": "Stmt1446575236270"
    }
  ],
  "Version": "2012-10-17"
}
EOF
}
```

It will receive the following error after upgrading:

```
│ Error: Value for unconfigurable attribute
│
│   with aws_s3_bucket.accesslogs_bucket,
│   on main.tf line 1, in resource "aws_s3_bucket" "accesslogs_bucket":
│    1: resource "aws_s3_bucket" "accesslogs_bucket" {
│
│ Can't configure a value for "policy": its value will be decided automatically based on the result of applying this configuration.
```

Since the `policy` argument changed to read-only, the recommendation is to update the configuration to use the `aws_s3_bucket_policy`
resource and remove any reference to `policy` in the `aws_s3_bucket` resource:

```terraform
resource "aws_s3_bucket" "accesslogs_bucket" {
  # ... other configuration ...
}

resource "aws_s3_bucket_policy" "example" {
  bucket = aws_s3_bucket.accesslogs_bucket.id
  policy = <<EOF
{
  "Id": "Policy1446577137248",
  "Statement": [
    {
      "Action": "s3:PutObject",
      "Effect": "Allow",
      "Principal": {
        "AWS": "${data.aws_elb_service_account.current.arn}"
      },
      "Resource": "arn:${data.aws_partition.current.partition}:s3:::example/*",
      "Sid": "Stmt1446575236270"
    }
  ],
  "Version": "2012-10-17"
}
EOF
}
```

It is then recommended running `terraform import` on each new resource to prevent data loss, e.g.

```shell
$ terraform import aws_s3_bucket_policy.example example
aws_s3_bucket_policy.example: Importing from ID "example"...
aws_s3_bucket_policy.example: Import prepared!
  Prepared aws_s3_bucket_policy for import
aws_s3_bucket_policy.example: Refreshing state... [id=example]

Import successful!

The resources that were imported are shown above. These resources are now in
your Terraform state and will henceforth be managed by Terraform.
```

### `replication_configuration` Argument deprecation

Switch your Terraform configuration to the `aws_s3_bucket_replication_configuration` resource instead.

For example, given this previous configuration:

```terraform
resource "aws_s3_bucket" "source" {
  provider = aws.central

  # ... other configuration ...

  replication_configuration {
    role = aws_iam_role.replication.arn
    rules {
      id     = "foobar"
      status = "Enabled"
      filter {
        tags = {}
      }
      destination {
        bucket        = aws_s3_bucket.destination.arn
        storage_class = "STANDARD"
        replication_time {
          status  = "Enabled"
          minutes = 15
        }
        metrics {
          status  = "Enabled"
          minutes = 15
        }
      }
    }
  }
}
```

It will receive the following error after upgrading:

```
│ Error: Value for unconfigurable attribute
│
│   with aws_s3_bucket.source,
│   on main.tf line 1, in resource "aws_s3_bucket" "source":
│    1: resource "aws_s3_bucket" "source" {
│
│ Can't configure a value for "replication_configuration": its value will be decided automatically based on the result of applying this configuration.
```

Since the `replication_configuration` argument changed to read-only, the recommendation is to update the configuration to use the `aws_s3_bucket_replication_configuration`
resource and remove any references to `replication_configuration` and its nested arguments in the `aws_s3_bucket` resource:

```terraform
resource "aws_s3_bucket" "source" {
  provider = aws.central
  # ... other configuration ...
}

resource "aws_s3_bucket_replication_configuration" "example" {
  bucket = aws_s3_bucket.source.id
  role   = aws_iam_role.replication.arn

  rule {
    id     = "foobar"
    status = "Enabled"

    filter {}

    delete_marker_replication {
      status = "Enabled"
    }

    destination {
      bucket        = aws_s3_bucket.destination.arn
      storage_class = "STANDARD"

      replication_time {
        status = "Enabled"
        time {
          minutes = 15
        }
      }

      metrics {
        status = "Enabled"
        event_threshold {
          minutes = 15
        }
      }
    }
  }
}
```

It is then recommended running `terraform import` on each new resource to prevent data loss, e.g.

```shell
$ terraform import aws_s3_bucket_replication_configuration.example example
aws_s3_bucket_replication_configuration.example: Importing from ID "example"...
aws_s3_bucket_replication_configuration.example: Import prepared!
  Prepared aws_s3_bucket_replication_configuration for import
aws_s3_bucket_replication_configuration.example: Refreshing state... [id=example]

Import successful!

The resources that were imported are shown above. These resources are now in
your Terraform state and will henceforth be managed by Terraform.
```

### `request_payer` Argument deprecation

Switch your Terraform configuration to the `aws_s3_bucket_request_payment_configuration` resource instead.

For example, given this previous configuration:

```terraform
resource "aws_s3_bucket" "example" {
  # ... other configuration ...
  request_payer = "Requester"
}
```

It will receive the following error after upgrading:

```
│ Error: Value for unconfigurable attribute
│
│   with aws_s3_bucket.example,
│   on main.tf line 1, in resource "aws_s3_bucket" "example":
│    1: resource "aws_s3_bucket" "example" {
│
│ Can't configure a value for "request_payer": its value will be decided automatically based on the result of applying this configuration.
```

Since the `request_payer` argument changed to read-only, the recommendation is to update the configuration to use the `aws_s3_bucket_request_payment_configuration`
resource and remove any reference to `request_payer` in the `aws_s3_bucket` resource:

```terraform
resource "aws_s3_bucket" "example" {
  # ... other configuration ...
}

resource "aws_s3_bucket_request_payment_configuration" "example" {
  bucket = aws_s3_bucket.example.id
  payer  = "Requester"
}
```

It is then recommended running `terraform import` on each new resource to prevent data loss, e.g.

```shell
$ terraform import aws_s3_bucket_request_payment_configuration.example example
aws_s3_bucket_request_payment_configuration.example: Importing from ID "example"...
aws_s3_bucket_request_payment_configuration.example: Import prepared!
  Prepared aws_s3_bucket_request_payment_configuration for import
aws_s3_bucket_request_payment_configuration.example: Refreshing state... [id=example]

Import successful!

The resources that were imported are shown above. These resources are now in
your Terraform state and will henceforth be managed by Terraform.
```

### `server_side_encryption_configuration` Argument deprecation

Switch your Terraform configuration to the `aws_s3_bucket_server_side_encryption_configuration` resource instead.

For example, given this previous configuration:

```terraform
resource "aws_s3_bucket" "example" {
  # ... other configuration ...
  server_side_encryption_configuration {
    rule {
      apply_server_side_encryption_by_default {
        kms_master_key_id = aws_kms_key.mykey.arn
        sse_algorithm     = "aws:kms"
      }
    }
  }
}
```

It will receive the following error after upgrading:

```
│ Error: Value for unconfigurable attribute
│
│   with aws_s3_bucket.example,
│   on main.tf line 1, in resource "aws_s3_bucket" "example":
│    1: resource "aws_s3_bucket" "example" {
│
│ Can't configure a value for "server_side_encryption_configuration": its value will be decided automatically based on the result of applying this configuration.
```

Since the `server_side_encryption_configuration` argument changed to read-only, the recommendation is to update the configuration to use the `aws_s3_bucket_server_side_encryption_configuration`
resource and remove any references to `server_side_encryption_configuration` and its nested arguments in the `aws_s3_bucket` resource:

```terraform
resource "aws_s3_bucket" "example" {
  # ... other configuration ...
}

resource "aws_s3_bucket_server_side_encryption_configuration" "example" {
  bucket = aws_s3_bucket.example.id

  rule {
    apply_server_side_encryption_by_default {
      kms_master_key_id = aws_kms_key.mykey.arn
      sse_algorithm     = "aws:kms"
    }
  }
}
```

It is then recommended running `terraform import` on each new resource to prevent data loss, e.g.

```shell
$ terraform import aws_s3_bucket_server_side_encryption_configuration.example example
aws_s3_bucket_server_side_encryption_configuration.example: Importing from ID "example"...
aws_s3_bucket_server_side_encryption_configuration.example: Import prepared!
  Prepared aws_s3_bucket_server_side_encryption_configuration for import
aws_s3_bucket_server_side_encryption_configuration.example: Refreshing state... [id=example]

Import successful!

The resources that were imported are shown above. These resources are now in
your Terraform state and will henceforth be managed by Terraform.
```

### `versioning` Argument deprecation

Switch your Terraform configuration to the `aws_s3_bucket_versioning` resource instead.

For example, given this previous configuration:

```terraform
resource "aws_s3_bucket" "example" {
  # ... other configuration ...
  versioning {
    enabled = true
  }
}
```

It will receive the following error after upgrading:

```
│ Error: Value for unconfigurable attribute
│
│   with aws_s3_bucket.example,
│   on main.tf line 1, in resource "aws_s3_bucket" "example":
│    1: resource "aws_s3_bucket" "example" {
│
│ Can't configure a value for "versioning": its value will be decided automatically based on the result of applying this configuration.
```

Since the `versioning` argument changed to read-only, the recommendation is to update the configuration to use the `aws_s3_bucket_versioning`
resource and remove any references to `versioning` and its nested arguments in the `aws_s3_bucket` resource:

```terraform
resource "aws_s3_bucket" "example" {
  # ... other configuration ...
}

resource "aws_s3_bucket_versioning" "example" {
  bucket = aws_s3_bucket.example.id
  versioning_configuration {
    status = "Enabled"
  }
}
```

It is then recommended running `terraform import` on each new resource to prevent data loss, e.g.

```shell
$ terraform import aws_s3_bucket_versioning.example example
aws_s3_bucket_versioning.example: Importing from ID "example"...
aws_s3_bucket_versioning.example: Import prepared!
  Prepared aws_s3_bucket_versioning for import
aws_s3_bucket_versioning.example: Refreshing state... [id=example]

Import successful!

The resources that were imported are shown above. These resources are now in
your Terraform state and will henceforth be managed by Terraform.
```

### `website`, `website_domain`, and `website_endpoint` Argument deprecation

Switch your Terraform configuration to the `aws_s3_bucket_website_configuration` resource instead.

For example, given this previous configuration:

```terraform
resource "aws_s3_bucket" "example" {
  # ... other configuration ...
  website {
    index_document = "index.html"
    error_document = "error.html"
  }
}
```

It will receive the following error after upgrading:

```
│ Error: Value for unconfigurable attribute
│
│   with aws_s3_bucket.example,
│   on main.tf line 1, in resource "aws_s3_bucket" "example":
│    1: resource "aws_s3_bucket" "example" {
│
│ Can't configure a value for "website": its value will be decided automatically based on the result of applying this configuration.
```

Since the `website` argument changed to read-only, the recommendation is to update the configuration to use the `aws_s3_bucket_website_configuration`
resource and remove any references to `website` and its nested arguments in the `aws_s3_bucket` resource:

```terraform
resource "aws_s3_bucket" "example" {
  # ... other configuration ...
}

resource "aws_s3_bucket_website_configuration" "example" {
  bucket = aws_s3_bucket.example.id

  index_document {
    suffix = "index.html"
  }

  error_document {
    key = "error.html"
  }
}
```

It is then recommended running `terraform import` on each new resource to prevent data loss, e.g.

```shell
$ terraform import aws_s3_bucket_website_configuration.example example
aws_s3_bucket_website_configuration.example: Importing from ID "example"...
aws_s3_bucket_website_configuration.example: Import prepared!
  Prepared aws_s3_bucket_website_configuration for import
aws_s3_bucket_website_configuration.example: Refreshing state... [id=example]

Import successful!

The resources that were imported are shown above. These resources are now in
your Terraform state and will henceforth be managed by Terraform.
```

For configurations that previously used the `website_domain` attribute to create Route 53 alias records e.g.

```terraform
resource "aws_route53_zone" "main" {
  name = "domain.test"
}

resource "aws_s3_bucket" "website" {
  # ... other configuration ...
  website {
    index_document = "index.html"
    error_document = "error.html"
  }
}

resource "aws_route53_record" "alias" {
  zone_id = aws_route53_zone.main.zone_id
  name    = "www"
  type    = "A"

  alias {
    zone_id                = aws_s3_bucket.website.hosted_zone_id
    name                   = aws_s3_bucket.website.website_domain
    evaluate_target_health = true
  }
}
```

An updated configuration:

```terraform
resource "aws_route53_zone" "main" {
  name = "domain.test"
}

resource "aws_s3_bucket" "website" {
  # ... other configuration ...
}

resource "aws_s3_bucket_website_configuration" "example" {
  bucket = aws_s3_bucket.website.id

  index_document {
    suffix = "index.html"
  }
}

resource "aws_route53_record" "alias" {
  zone_id = aws_route53_zone.main.zone_id
  name    = "www"
  type    = "A"

  alias {
    zone_id                = aws_s3_bucket.website.hosted_zone_id
    name                   = aws_s3_bucket_website_configuration.example.website_domain
    evaluate_target_health = true
  }
}
```

## Resource: aws_s3_bucket_object

The `aws_s3_bucket_object` resource is deprecated and will be removed in a future version. Use `aws_s3_object` instead, where new features and fixes will be added.

When replacing `aws_s3_bucket_object` with `aws_s3_object` in your configuration, on the next apply, Terraform will recreate the object. If you prefer to not have Terraform recreate the object, import the object using `aws_s3_object`.

For example, the following will import an S3 object into state, assuming the configuration exists, as `aws_s3_object.example`:

```console
% terraform import aws_s3_object.example s3://some-bucket-name/some/key.txt
```

## Resource: aws_spot_instance_request

### instance_interruption_behaviour Argument removal

Switch your Terraform configuration to the `engine_version_actual` attribute instead.

For example, given this previous configuration:

```terraform
resource "aws_spot_instance_request" "example" {
  # ... other configuration ...
  instance_interruption_behaviour = "hibernate"
}
```

An updated configuration:

```terraform
resource "aws_spot_instance_request" "example" {
  # ... other configuration ...
  instance_interruption_behavior =  "hibernate"
}
```

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
