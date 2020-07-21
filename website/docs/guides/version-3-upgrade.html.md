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

~> **NOTE:** Version 3.0.0 and later of the AWS Provider can only be automatically installed on Terraform 0.12 and later.

Upgrade topics:

<!-- TOC depthFrom:2 depthTo:2 -->

- [Provider Version Configuration](#provider-version-configuration)
- [Provider Authentication Updates](#provider-authentication-updates)
- [Data Source: aws_availability_zones](#data-source-aws_availability_zones)
- [Data Source: aws_lambda_invocation](#data-source-aws_lambda_invocation)
- [Resource: aws_acm_certificate](#resource-aws_acm_certificate)
- [Resource: aws_api_gateway_method_settings](#resource-aws_api_gateway_method_settings)
- [Resource: aws_autoscaling_group](#resource-aws_autoscaling_group)
- [Resource: aws_dx_gateway](#resource-aws_dx_gateway)
- [Resource: aws_elastic_transcoder_preset](#resource-aws_elastic_transcoder_preset)
- [Resource: aws_emr_cluster](#resource-aws_emr_cluster)
- [Resource: aws_lambda_alias](#resource-aws_lambda_alias)
- [Resource: aws_launch_template](#resource-aws_launch_template)
- [Resource: aws_lb_listener_rule](#resource-aws_lb_listener_rule)
- [Resource: aws_msk_cluster](#resource-aws_msk_cluster)
- [Resource: aws_rds_cluster](#resource-aws_rds_cluster)
- [Resource: aws_s3_bucket](#resource-aws_s3_bucket)
- [Resource: aws_security_group](#resource-aws_security_group)
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

## Resource: aws_acm_certificate

### subject_alternative_names Changed from List to Set

Previously the `subject_alternative_names` argument was stored in the Terraform state as an ordered list while the API returned information in an unordered manner. The attribute is now configured as a set instead of a list. Certain Terraform configuration language features distinguish between these two attribute types such as not being able to index a set (e.g. `aws_acm_certificate.example.subject_alternative_names[0]` is no longer a valid reference). Depending on the implementation details of a particular configuration using `subject_alternative_names` as a reference, possible solutions include changing references to using `for`/`for_each` or using the `tolist()` function as a temporary workaround to keep the previous behavior until an appropriate configuration (properly using the unordered set) can be determined. Usage questions can be submitted to the [community forums](https://discuss.hashicorp.com/c/terraform-providers/tf-aws/33).

### certificate_body, certificate_chain, and private_key Arguments No Longer Stored as Hash

Previously when the `certificate_body`, `certificate_chain`, and `private_key` arguments were stored in state, they were stored as a hash of the actual value. This prevented Terraform from properly updating the resource when necessary and the hashing has been removed. The Terraform AWS Provider will show an update to these arguments on the first apply after upgrading to version 3.0.0, which is fixing the Terraform state to remove the hash. Since the `private_key` attribute is marked as sensitive, the values in the update will not be visible in the Terraform output. If the non-hashed values have not changed, then no update is occurring other than the Terraform state update. If these arguments are the only updates and they all match the hash removal, the apply will occur without submitting API calls.

## Resource: aws_api_gateway_method_settings

### throttling_burst_limit and throttling_rate_limit Arguments Now Default to -1

Previously when the `throttling_burst_limit` or `throttling_rate_limit` argument was not configured, the resource would enable throttling and set the limit value to the AWS API Gateway default. In addition, as these arguments were marked as `Computed`, Terraform ignored any subsequent changes made to these arguments in the resource. These behaviors have been removed and, by default, the `throttling_burst_limit` and `throttling_rate_limit` arguments will be disabled in the resource with a value of `-1`.

## Resource: aws_autoscaling_group

### availability_zones and vpc_zone_identifier Arguments Now Report Plan-Time Conflict

Specifying both the `availability_zones` and `vpc_zone_identifier` arguments previously led to confusing behavior and errors. Now this issue is reported at plan-time. Use the `null` value instead of `[]` (empty list) in conditionals to ensure this validation does not unexpectedly trigger.

### Drift detection enabled for `load_balancers` and `target_group_arns` arguments

If you previously set one of these arguments to an empty list to enable drift detection (e.g. when migrating an ASG from ELB to ALB), this can be updated as follows.

For example, given this previous configuration:

```hcl
resource "aws_autoscaling_group" "example" {
  # ... other configuration ...
  load_balancers = []
  target_group_arns = [aws_lb_target_group.example.arn]
}
```

An updated configuration:

```hcl
resource "aws_autoscaling_group" "example" {
  # ... other configuration ...
  target_group_arns = [aws_lb_target_group.example.arn]
}
```

If `aws_autoscaling_attachment` resources reference your ASG configurations, you will need to add the [`lifecycle` configuration block](/docs/configuration/resources.html#lifecycle-lifecycle-customizations) with an `ignore_changes` argument to prevent Terraform non-empty plans (i.e. forcing resource update) during the next state refresh.

For example, given this previous configuration:

```hcl
resource "aws_autoscaling_attachment" "example" {
  autoscaling_group_name = aws_autoscaling_group.example.id
  elb                    = aws_elb.example.id
}

resource "aws_autoscaling_group" "example"{
  # ... other configuration ...
}
```

An updated configuration:

```hcl
resource "aws_autoscaling_attachment" "example" {
  autoscaling_group_name = aws_autoscaling_group.example.id
  elb                    = aws_elb.example.id
}

resource "aws_autoscaling_group" "example"{
  # ... other configuration ...

  lifecycle {
    ignore_changes = [load_balancers, target_group_arns]
  }
}
```

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

## Resource: aws_lambda_alias

### Import No Longer Converts Function Name to ARN

Previously the resource import would always convert the `function_name` portion of the import identifier into the ARN format. Configurations using the Lambda Function name would show this as an unexpected difference after import. Now this will passthrough the given value on import whether its a Lambda Function name or ARN.

## Resource: aws_launch_template

### network_interfaces.delete_on_termination Argument type change

The `network_interfaces.delete_on_termination` argument is now of type `string`, allowing an unspecified value for the argument since the previous `bool` type only allowed for `true/false` and defaulted to `false` when no value was set. Now to enforce `delete_on_termination` to `false`, the string `"false"` or bare `false` value must be used.

For example, given this previous configuration:

```hcl
resource "aws_launch_template" "example" {
  # ... other configuration ...

  network_interfaces {
    # ... other configuration ...

    delete_on_termination = null
  }
}
```

An updated configuration:

```hcl
resource "aws_launch_template" "example" {
  # ... other configuration ...

  network_interfaces {
    # ... other configuration ...

    delete_on_termination = false
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

## Resource: aws_msk_cluster

### encryption_info.encryption_in_transit.client_broker Default Updated to Match API

A few weeks after general availability launch and initial release of the `aws_msk_cluster` resource, the MSK API default for client broker encryption switched from `TLS_PLAINTEXT` to `TLS`. The attribute default has now been updated to match the more secure API default, however existing Terraform configurations may show a difference if this setting is not configured.

To continue using the old default when it was previously not configured, add or modify this configuration:

```hcl
resource "aws_msk_cluster" "example" {
  # ... other configuration ...

  encryption_info {
    # ... potentially other configuration ...

    encryption_in_transit {
      # ... potentially other configuration ...

      client_broker = "TLS_PLAINTEXT"
    }
  }
}
```

## Resource: aws_rds_cluster

### scaling_configuration.min_capacity Now Defaults to 1

Previously when the `min_capacity` argument in a `scaling_configuration` block was not configured, the resource would default to 2. This behavior has been updated to align with the AWS RDS Cluster API default of 1. 

## Resource: aws_s3_bucket

### Removal of Automatic aws_s3_bucket_policy Import

Previously when importing the `aws_s3_bucket` resource with the [`terraform import` command](/docs/commands/import.html), the Terraform AWS Provider would automatically attempt to import an associated `aws_s3_bucket_policy` resource as well. This automatic resource import has been removed. Use the [`aws_s3_bucket_policy` resource import](/docs/providers/aws/r/s3_bucket_policy.html#import) to import that resource separately.

### region Attribute Is Now Read-Only

The `region` attribute is no longer configurable, but it remains as a read-only attribute. The region of the `aws_s3_bucket` resource is determined by the region of the Terraform AWS Provider, similar to all other resources.

For example, given this previous configuration:

```hcl
resource "aws_s3_bucket" "example" {
  # ... other configuration ...

  region = "us-west-2"
}
```

An updated configuration:

```hcl
resource "aws_s3_bucket" "example" {
  # ... other configuration ...
}
```

## Resource: aws_security_group

### Removal of Automatic aws_security_group_rule Import

Previously when importing the `aws_security_group` resource with the [`terraform import` command](/docs/commands/import.html), the Terraform AWS Provider would automatically attempt to import an associated `aws_security_group_rule` resource(s) as well. This automatic resource import has been removed. Use the [`aws_security_group_rule` resource import](/docs/providers/aws/r/security_group_rule.html#import) to import those resources separately.

## Resource: aws_sns_platform_application

### platform_credential and platform_principal Arguments No Longer Stored as SHA256 Hash

Previously when the `platform_credential` and `platform_principal` arguments were stored in state, they were stored as a SHA256 hash of the actual value. This prevented Terraform from properly updating the resource when necessary and the hashing has been removed. The Terraform AWS Provider will show an update to these arguments on the first apply after upgrading to version 3.0.0, which is fixing the Terraform state to remove the hash. Since the attributes are marked as sensitive, the values in the update will not be visible in the Terraform output. If the non-hashed values have not changed, then no update is occurring other than the Terraform state update. If these arguments are the only two updates and they both match the SHA256 removal, the apply will occur without submitting an actual `SetPlatformApplicationAttributes` API call.

## Resource: aws_spot_fleet_request

### valid_until Argument No Longer Uses 24 Hour Default

Previously when the `valid_until` argument was not configured, the resource would default to a 24 hour request. This behavior has been removed and allows for non-expiring requests. To recreate the old behavior, the [`time_offset` resource](/docs/providers/time/r/offset.html) can potentially be used.
