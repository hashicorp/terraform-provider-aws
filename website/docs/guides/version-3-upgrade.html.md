---
subcategory: ""
layout: "aws"
page_title: "Terraform AWS Provider Version 3 Upgrade Guide"
description: |-
  Terraform AWS Provider Version 3 Upgrade Guide
---

# Terraform AWS Provider Version 3 Upgrade Guide

Version 3.0.0 of the AWS provider for Terraform is a major release and includes some changes that you will need to consider when upgrading. This guide is intended to help with that process and focuses only on changes from version 2.X to version 3.0.0. See the [Version 2 Upgrade Guide](/docs/providers/aws/guides/version-2-upgrade.html) for information about upgrading from 1.X to version 2.0.0.

Most of the changes outlined in this guide have been previously marked as deprecated in the Terraform plan/apply output throughout previous provider releases. These changes, such as deprecation notices, can always be found in the [Terraform AWS Provider CHANGELOG](https://github.com/terraform-providers/terraform-provider-aws/blob/master/CHANGELOG.md).

Upgrade topics:

<!-- TOC depthFrom:2 depthTo:2 -->

- [Provider Version Configuration](#provider-version-configuration)
- [Provider Authentication Updates](#provider-authentication-updates)
- [Data Source: aws_availability_zones](#data-source-aws_availability_zones)
- [Data Source: aws_lambda_invocation](#data-source-aws_lambda_invocation)
- [Resource: aws_autoscaling_group](#resource-aws_autoscaling_group)
- [Resource: aws_dx_gateway](#resource-aws_dx_gateway)
- [Resource: aws_elastic_transcoder_preset](#resource-aws_elastic_transcoder_preset)
- [Resource: aws_emr_cluster](#resource-aws_emr_cluster)
- [Resource: aws_lb_listener_rule](#resource-aws_lb_listener_rule)
- [Resource: aws_s3_bucket](#resource-aws_s3_bucket)
- [Resource: aws_sns_platform_application](#resource-aws_sns_platform_application)
- [Resource: aws_spot_fleet_request](#resource-aws_spot_fleet_request)

<!-- /TOC -->

## Provider Version Configuration

-> Before upgrading to version 3.0.0, it is recommended to upgrade to the most recent 2.X version of the provider and ensure that your environment successfully runs [`terraform plan`](https://www.terraform.io/docs/commands/plan.html) without unexpected changes or deprecation notices.

It is recommended to use [version constraints when configuring Terraform providers](https://www.terraform.io/docs/configuration/providers.html#provider-versions). If you are following that recommendation, update the version constraints in your Terraform configuration and run [`terraform init`](https://www.terraform.io/docs/commands/init.html) to download the new version.

For example, given this previous configuration:

```hcl
provider "aws" {
  # ... other configuration ...

  version = "~> 2.70"
}
```

Update to latest 3.X version:

```hcl
provider "aws" {
  # ... other configuration ...

  version = "~> 3.0"
}
```

## Provider Authentication Updates

### Authentication Ordering

Previously, the provider preferred credentials in the following order:

- Static credentials (those defined in the Terraform configuration)
- Environment variables (e.g. `AWS_ACCESS_KEY_ID` or `AWS_PROFILE`)
- Shared credentials file (e.g. `~/.aws/credentials`)
- EC2 Instance Metadata Service
- Default AWS Go SDK handling (shared configuration, CodeBuild/ECS/EKS)

The provider now prefers the following credential ordering:

- Static credentials (those defined in the Terraform configuration)
- Environment variables (e.g. `AWS_ACCESS_KEY_ID` or `AWS_PROFILE`)
- Shared credentials and/or configuration file (e.g. `~/.aws/credentials` and `~/.aws/config`)
- Default AWS Go SDK handling (shared configuration, CodeBuild/ECS/EKS, EC2 Instance Metadata Service)

This means workarounds of disabling the EC2 Instance Metadata Service handling to enable CodeBuild/ECS/EKS credentials or to enable other credential methods such as `credential_process` in the AWS shared configuration are no longer necessary.

### Shared Configuration File Automatically Enabled

The `AWS_SDK_LOAD_CONFIG` environment variable is no longer necessary for the provider to automatically load the AWS shared configuration file (e.g. `~/.aws/config`).

### Removal of AWS_METADATA_TIMEOUT Environment Variable Usage

The provider now relies on the default AWS Go SDK timeouts for interacting with the EC2 Instance Metadata Service.

## Data Source: aws_availability_zones

### blacklisted_names Attribute Removal

Switch your Terraform configuration to the `exclude_names` attribute instead.

For example, given this previous configuration:

```hcl
data "aws_availability_zones" "example" {
  blacklisted_names = ["us-west-2d"]
}
```

An updated configuration:

```hcl
data "aws_availability_zones" "example" {
  exclude_names = ["us-west-2d"]
}
```

### blacklisted_zone_ids Attribute Removal

Switch your Terraform configuration to the `exclude_zone_ids` attribute instead.

For example, given this previous configuration:

```hcl
data "aws_availability_zones" "example" {
  blacklisted_zone_ids = ["usw2-az4"]
}
```

An updated configuration:

```hcl
data "aws_availability_zones" "example" {
  exclude_zone_ids = ["usw2-az4"]
}
```

## Data Source: aws_lambda_invocation

### result_map Attribute Removal

Switch your Terraform configuration to the `result` attribute with the [`jsondecode()` function](https://www.terraform.io/docs/configuration/functions/jsondecode.html) instead.

For example, given this previous configuration:

```hcl
# In Terraform 0.11 and earlier, the result_map attribute can be used
# to convert a result JSON string to a map of string keys to string values.
output "lambda_result" {
  value = "${data.aws_lambda_invocation.example.result_map["key1"]}"
}
```

An updated configuration:

```hcl
# In Terraform 0.12 and later, the jsondecode() function can be used
# to convert a result JSON string to native Terraform types.
output "lambda_result" {
  value = jsondecode(data.aws_lambda_invocation.example.result)["key1"]
}
```

## Resource: aws_autoscaling_group

### availability_zones and vpc_zone_identifier Arguments Now Report Plan-Time Conflict

Specifying both the `availability_zones` and `vpc_zone_identifier` arguments previously led to confusing behavior and errors. Now this issue is reported at plan-time. Use the `null` value instead of `[]` (empty list) in conditionals to ensure this validation does not unexpectedly trigger.

## Resource: aws_dx_gateway

### Removal of Automatic aws_dx_gateway_association Import

Previously when importing the `aws_dx_gateway` resource with the [`terraform import` command](/docs/commands/import.html), the Terraform AWS Provider would automatically attempt to import an associated `aws_dx_gateway_association` resource(s) as well. This automatic resource import has been removed. Use the [`aws_dx_gateway_association` resource import](/docs/providers/aws/r/dx_gateway_association.html#import) to import those resources separately.

## Resource: aws_elastic_transcoder_preset

### video Configuration Block max_frame_rate Argument No Longer Uses 30 Default

Previously when the `max_frame_rate` argument was not configured, the resource would default to 30. This behavior has been removed and allows for auto frame rate presets to automatically set the appropriate value.

## Resource: aws_emr_cluster

### core_instance_count Argument Removal

Switch your Terraform configuration to the `core_instance_group` configuration block instead.

For example, given this previous configuration:

```hcl
resource "aws_emr_cluster" "example" {
  # ... other configuration ...

  core_instance_count = 2
}
```

An updated configuration:

```hcl
resource "aws_emr_cluster" "example" {
  # ... other configuration ...

  core_instance_group {
    # ... other configuration ...

    instance_count = 2
  }
}
```

### core_instance_type Argument Removal

Switch your Terraform configuration to the `core_instance_group` configuration block instead.

For example, given this previous configuration:

```hcl
resource "aws_emr_cluster" "example" {
  # ... other configuration ...

  core_instance_type = "m4.large"
}
```

An updated configuration:

```hcl
resource "aws_emr_cluster" "example" {
  # ... other configuration ...

  core_instance_group {
    instance_type = "m4.large"
  }
}
```

### instance_group Configuration Block Removal

Switch your Terraform configuration to the `master_instance_group` and `core_instance_group` configuration blocks instead. For any task instance groups, use the `aws_emr_instance_group` resource.

For example, given this previous configuration:

```hcl
resource "aws_emr_cluster" "example" {
  # ... other configuration ...

  instance_group {
    instance_role  = "MASTER"
    instance_type  = "m4.large"
  }

  instance_group {
    instance_count = 1
    instance_role  = "CORE"
    instance_type  = "c4.large"
  }

  instance_group {
    instance_count = 2
    instance_role  = "TASK"
    instance_type  = "c4.xlarge"
  }
}
```

An updated configuration:

```hcl
resource "aws_emr_cluster" "example" {
  # ... other configuration ...

  master_instance_group {
    instance_type = "m4.large"
  }

  core_instance_group {
    instance_count = 1
    instance_type  = "c4.large"
  }
}

resource "aws_emr_instance_group" "example" {
  cluster_id     = "${aws_emr_cluster.example.id}"
  instance_count = 2
  instance_type  = "c4.xlarge"
}
```

### master_instance_type Argument Removal

Switch your Terraform configuration to the `master_instance_group` configuration block instead.

For example, given this previous configuration:

```hcl
resource "aws_emr_cluster" "example" {
  # ... other configuration ...

  master_instance_type = "m4.large"
}
```

An updated configuration:

```hcl
resource "aws_emr_cluster" "example" {
  # ... other configuration ...

  master_instance_group {
    instance_type = "m4.large"
  }
}
```

## Resource: aws_lb_listener_rule

### condition.field and condition.values Arguments Removal

Switch your Terraform configuration to use the `host_header` or `path_pattern` configuration block instead.

For example, given this previous configuration:

```hcl
resource "aws_lb_listener_rule" "example" {
  # ... other configuration ...

  condition {
    field  = "path-pattern"
    values = ["/static/*"]
  }
}
```

An updated configuration:

```hcl
resource "aws_lb_listener_rule" "example" {
  # ... other configuration ...

  condition {
    path_pattern {
      values = ["/static/*"]
    }
  }
}
```

## Resource: aws_s3_bucket

### Removal of Automatic aws_s3_bucket_policy Import

Previously when importing the `aws_s3_bucket` resource with the [`terraform import` command](/docs/commands/import.html), the Terraform AWS Provider would automatically attempt to import an associated `aws_s3_bucket_policy` resource as well. This automatic resource import has been removed. Use the [`aws_s3_bucket_policy` resource import](/docs/providers/aws/r/s3_bucket_policy.html#import) to import that resource separately.

## Resource: aws_sns_platform_application

### platform_credential and platform_principal Arguments No Longer Stored as SHA256 Hash

Previously when the `platform_credential` and `platform_principal` arguments were stored in state, they were stored as a SHA256 hash of the actual value. This prevented Terraform from properly updating the resource when necessary and the hashing has been removed. The Terraform AWS Provider will show an update to these arguments on the first apply after upgrading to version 3.0.0, which is fixing the Terraform state to remove the hash. Since the attributes are marked as sensitive, the values in the update will not be visible in the Terraform output. If the non-hashed values have not changed, then no update is occurring other than the Terraform state update. If these arguments are the only two updates and they both match the SHA256 removal, the apply will occur without submitting an actual `SetPlatformApplicationAttributes` API call.

## Resource: aws_spot_fleet_request

### valid_until Argument No Longer Uses 24 Hour Default

Previously when the `valid_until` argument was not configured, the resource would default to a 24 hour request. This behavior has been removed and allows for non-expiring requests. To recreate the old behavior, the [`time_offset` resource](/docs/providers/time/r/offset.html) can potentially be used.
