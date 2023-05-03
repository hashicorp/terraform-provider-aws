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
