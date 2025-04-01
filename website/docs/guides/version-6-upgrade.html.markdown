---
subcategory: ""
layout: "aws"
page_title: "Terraform AWS Provider Version 6 Upgrade Guide"
description: |-
  Terraform AWS Provider Version 6 Upgrade Guide
---

# Terraform AWS Provider Version 6 Upgrade Guide

Version 6.0.0 of the AWS provider for Terraform is a major release and includes changes that you need to consider when upgrading. This guide will help with that process and focuses only on changes from version 5.x to version 6.0.0. See the [Version 5 Upgrade Guide](/docs/providers/aws/guides/version-5-upgrade.html) for information on upgrading from 4.x to version 5.0.0.

Upgrade topics:

<!-- TOC depthFrom:2 depthTo:2 -->

- [Provider Version Configuration](#provider-version-configuration)
- [Provider Arguments](#provider-arguments)
- [Dropping Support For Amazon SimpleDB](#dropping-support-for-amazon-simpledb)
- [Dropping Support For Amazon Worklink](#dropping-support-for-amazon-worklink)
- [AWS OpsWorks Stacks End of Life](#aws-opsworks-stacks-end-of-life)
- [resource/aws_redshift_cluster](#resourceaws_redshift_cluster)
- [resource/aws_redshift_service_account](#resourceaws_redshift_service_account)

<!-- /TOC -->

## Provider Version Configuration

-> Before upgrading to version 6.0.0, upgrade to the most recent 5.X version of the provider and ensure that your environment successfully runs [`terraform plan`](https://www.terraform.io/docs/commands/plan.html). You should not see changes you don't expect, or deprecation notices for anything mentioned in this guide.

Use [version constraints when configuring Terraform providers](https://www.terraform.io/docs/configuration/providers.html#provider-versions). If you are following that recommendation, update the version constraints in your Terraform configuration and run [`terraform init -upgrade`](https://www.terraform.io/docs/commands/init.html) to download the new version.

For example, given this previous configuration:

```terraform
terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.92"
    }
  }
}

provider "aws" {
  # Configuration options
}
```

Update to the latest 6.X version:

```terraform
terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 6.0"
    }
  }
}

provider "aws" {
  # Configuration options
}
```

## Provider Arguments

Version 6.0.0 removes these `provider` arguments:

* `endpoints.opsworks` - Removed following AWS OpsWorks Stacks End of Life
* `endpoints.simpledb` and `endpoints.sdb` - Removed following dropping support for Amazon SimpleDB
* `endpoints.worklink` - Removed following dropping support for Amazon Worklink

## Dropping Support For Amazon SimpleDB

As the [AWS SDK for Go v2](https://docs.aws.amazon.com/sdk-for-go/v2/developer-guide/welcome.html) does not support Amazon SimpleDB, the `aws_simpledb_domain` resource has been removed.

## Dropping Support For Amazon Worklink

As the [AWS SDK for Go v2](https://docs.aws.amazon.com/sdk-for-go/v2/developer-guide/welcome.html) has [dropped support](https://github.com/aws/aws-sdk-go-v2/pull/2814) for Amazon Worklink, the following resources have been removed:

* `aws_worklink_fleet`
* `aws_worklink_website_certificate_authority_association`

## AWS OpsWorks Stacks End of Life

As the AWS OpsWorks Stacks service has reached [End Of Life](https://docs.aws.amazon.com/opsworks/latest/userguide/stacks-eol-faqs.html), the following resources have been removed:

* `aws_opsworks_application`
* `aws_opsworks_custom_layer`
* `aws_opsworks_ecs_cluster_layer`
* `aws_opsworks_ganglia_layer`
* `aws_opsworks_haproxy_layer`
* `aws_opsworks_instance`
* `aws_opsworks_java_app_layer`
* `aws_opsworks_memcached_layer`
* `aws_opsworks_mysql_layer`
* `aws_opsworks_nodejs_app_layer`
* `aws_opsworks_permission`
* `aws_opsworks_php_app_layer`
* `aws_opsworks_rails_app_layer`
* `aws_opsworks_rds_db_instance`
* `aws_opsworks_stack`
* `aws_opsworks_static_web_layer`
* `aws_opsworks_user_profile`

## resource/aws_redshift_cluster

* The `publicly_accessible` attribute now defaults to `false`.
* Remove `snapshot_copy` from configuration as it no longer exists. Use the `aws_redshift_snapshot_copy` resource instead.
* Remove `logging` from configuration as it no longer exists. Use the `aws_redshift_logging` resource instead.

## resource/aws_redshift_service_account

The `aws_redshift_service_account` resource has been removed. AWS [recommends](https://docs.aws.amazon.com/redshift/latest/mgmt/db-auditing.html#db-auditing-bucket-permissions) that a [service principal name](https://docs.aws.amazon.com/IAM/latest/UserGuide/reference_policies_elements_principal.html#principal-services) should be used instead of an AWS account ID in any relevant IAM policy.

## resource/aws_spot_instance_request

* Remove `block_duration_minutes` from configuration as it no longer exists.
