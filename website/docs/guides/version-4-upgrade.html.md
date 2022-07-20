---
subcategory: ""
layout: "aws"
page_title: "Terraform AWS Provider Version 4 Upgrade Guide"
description: |-
  Terraform AWS Provider Version 4 Upgrade Guide
---

# Terraform AWS Provider Version 4 Upgrade Guide

Version 4.0.0 of the AWS provider for Terraform is a major release and includes some changes that you will need to consider when upgrading. We intend this guide to help with that process and focus only on changes from version 3.X to version 4.0.0. See the [Version 3 Upgrade Guide](/docs/providers/aws/guides/version-3-upgrade.html) for information about upgrading from 1.X to version 3.0.0.

We previously marked most of the changes we outline in this guide as deprecated in the Terraform plan/apply output throughout previous provider releases. You can find these changes, including deprecation notices, in the [Terraform AWS Provider CHANGELOG](https://github.com/hashicorp/terraform-provider-aws/blob/main/CHANGELOG.md).

~> **NOTE:** Versions 4.0.0 through v4.8.0 of the AWS Provider introduce significant breaking changes to the `aws_s3_bucket` resource. See [S3 Bucket Refactor](#s3-bucket-refactor) for more details.
We recommend upgrading to v4.9.0 or later of the AWS Provider instead, where only non-breaking changes and deprecation notices are introduced to the `aws_s3_bucket`. See  [Changes to S3 Bucket Drift Detection](#changes-to-s3-bucket-drift-detection) for additional considerations when upgrading to v4.9.0 or later.

~> **NOTE:** Version 4.0.0 of the AWS Provider introduces changes to the precedence of some authentication and configuration parameters.
These changes bring the provider in line with the AWS CLI and SDKs.
See [Changes to Authentication](#changes-to-authentication) for more details.

~> **NOTE:** Version 4.0.0 of the AWS Provider will be the last major version to support [EC2-Classic resources](#ec2-classic-resource-and-data-source-support) as AWS plans to fully retire EC2-Classic Networking. See the [AWS News Blog](https://aws.amazon.com/blogs/aws/ec2-classic-is-retiring-heres-how-to-prepare/) for additional details.

~> **NOTE:** Version 4.0.0 and 4.x.x versions of the AWS Provider will be the last versions compatible with Terraform 0.12-0.15.

Upgrade topics:

<!-- TOC depthFrom:2 depthTo:2 -->

- [Provider Version Configuration](#provider-version-configuration)
- [Changes to Authentication](#changes-to-authentication)
- [New Provider Arguments](#new-provider-arguments)
- [Changes to S3 Bucket Drift Detection](#changes-to-s3-bucket-drift-detection) (**Applicable to v4.9.0 and later of the AWS Provider**)
- [S3 Bucket Refactor](#s3-bucket-refactor) (**Only applicable to v4.0.0 through v4.8.0 of the AWS Provider**)
    - [`acceleration_status` Argument](#acceleration_status-argument)
    - [`acl` Argument](#acl-argument)
    - [`cors_rule` Argument](#cors_rule-argument)
    - [`grant` Argument](#grant-argument)
    - [`lifecycle_rule` Argument](#lifecycle_rule-argument)
    - [`logging` Argument](#logging-argument)
    - [`object_lock_configuration` `rule` Argument](#object_lock_configuration-rule-argument)
    - [`policy` Argument](#policy-argument)
    - [`replication_configuration` Argument](#replication_configuration-argument)
    - [`request_payer` Argument](#request_payer-argument)
    - [`server_side_encryption_configuration` Argument](#server_side_encryption_configuration-argument)
    - [`versioning` Argument](#versioning-argument)
    - [`website`, `website_domain`, and `website_endpoint` Arguments](#website-website_domain-and-website_endpoint-arguments)
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
- [Resource: aws_fsx_ontap_storage_virtual_machine](#resource-aws_fsx_ontap_storage_virtual_machine)
- [Resource: aws_lb_target_group](#resource-aws_lb_target_group)
- [Resource: aws_s3_bucket_object](#resource-aws_s3_bucket_object)
- [Resource: aws_spot_instance_request](#resource-aws_spot_instance_request)

<!-- /TOC -->

Additional Topics:

<!-- TOC depthFrom:2 depthTo:2 -->

- [EC2-Classic resource and data source support](#ec2-classic-resource-and-data-source-support)

<!-- /TOC -->


## Provider Version Configuration

-> Before upgrading to version 4.0.0, upgrade to the most recent 3.X version of the provider and ensure that your environment successfully runs [`terraform plan`](https://www.terraform.io/docs/commands/plan.html). You should not see changes you don't expect or deprecation notices.

Use [version constraints when configuring Terraform providers](https://www.terraform.io/docs/configuration/providers.html#provider-versions). If you are following that recommendation, update the version constraints in your Terraform configuration and run [`terraform init -upgrade`](https://www.terraform.io/docs/commands/init.html) to download the new version.

For example, given this previous configuration:

```terraform
terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 3.74"
    }
  }
}

provider "aws" {
  # Configuration options
}
```

Update to the latest 4.X version:

```terraform
terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 4.0"
    }
  }
}

provider "aws" {
  # Configuration options
}
```

## Changes to Authentication

The authentication configuration for the AWS Provider has changed in this version to match the behavior of other AWS products, including the AWS SDK and AWS CLI. _This will cause authentication failures in AWS provider configurations where you set a non-empty `profile` in the `provider` configuration but the profile does not correspond to an AWS profile with valid credentials._

Precedence for authentication settings is as follows:

* `provider` configuration
* Environment variables
* Shared credentials and configuration files (_e.g._, `~/.aws/credentials` and `~/.aws/config`)

In previous versions of the provider, you could explicitly set `profile` in the `provider`, and if the profile did not correspond to valid credentials, the provider would use credentials from environment variables. Starting in v4.0, the Terraform AWS provider enforces the precedence shown above, similarly to how the AWS SDK and AWS CLI behave.

In other words, when you explicitly set `profile` in `provider`, the AWS provider will not use environment variables per the precedence shown above. Before v4.0, if `profile` was configured in the `provider` configuration but did not correspond to an AWS profile or valid credentials, the provider would attempt to use environment variables. **This is no longer the case.** An explicitly set profile that does not have valid credentials will cause an authentication error.

For example, with the following, the environment variables will not be used:

```console
$ export AWS_ACCESS_KEY_ID="anaccesskey"
$ export AWS_SECRET_ACCESS_KEY="asecretkey"
```

```terraform
provider "aws" {
  region  = "us-west-2"
  profile = "customprofile"
}
```

## New Provider Arguments

Version 4.x adds these new `provider` arguments:

* `assume_role.duration` - Assume role duration as a string, _e.g._, `"1h"` or `"1h30s"`. Terraform AWS Provider v4.0.0 deprecates `assume_role.duration_seconds` and a future version will remove it.
* `custom_ca_bundle` - File containing custom root and intermediate certificates. Can also be configured using the `AWS_CA_BUNDLE` environment variable. (Setting `ca_bundle` in the shared config file is not supported.)
* `ec2_metadata_service_endpoint` - Address of the EC2 metadata service (IMDS) endpoint to use. Can also be set with the `AWS_EC2_METADATA_SERVICE_ENDPOINT` environment variable.
* `ec2_metadata_service_endpoint_mode` - Mode to use in communicating with the metadata service. Valid values are `IPv4` and `IPv6`. Can also be set with the `AWS_EC2_METADATA_SERVICE_ENDPOINT_MODE` environment variable.
* `s3_use_path_style` - Replaces `s3_force_path_style`, which has been deprecated in Terraform AWS Provider v4.0.0 and support will be removed in a future version.
* `shared_config_files` - List of paths to AWS shared config files. If not set, the default is `[~/.aws/config]`. A single value can also be set with the `AWS_CONFIG_FILE` environment variable.
* `shared_credentials_files` - List of paths to the shared credentials file. If not set, the default  is `[~/.aws/credentials]`. A single value can also be set with the `AWS_SHARED_CREDENTIALS_FILE` environment variable. Replaces `shared_credentials_file`, which has been deprecated in Terraform AWS Provider v4.0.0 and support will be removed in a future version.
* `sts_region` - Region where AWS STS operations will take place. For example, `us-east-1` and `us-west-2`.
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

## Changes to S3 Bucket Drift Detection

~> **NOTE:** This only applies to v4.9.0 and later of the AWS Provider.

~> **NOTE:** If you are migrating from v3.75.x of the AWS Provider and you have already adopted the standalone S3 bucket resources (e.g. `aws_s3_bucket_lifecycle_configuration`),
a [`lifecycle` configuration block to ignore changes](https://www.terraform.io/language/meta-arguments/lifecycle#ignore_changes) to the internal parameters of the source `aws_s3_bucket` resources will no longer be necessary and can be removed upon upgrade.

~> **NOTE:** In the next major version, v5.0, the parameters listed below will be removed entirely from the `aws_s3_bucket` resource.
For this reason, a deprecation notice is printed in the Terraform CLI for each of the parameters when used in a configuration.

To remediate the breaking changes introduced to the `aws_s3_bucket` resource in v4.0.0 of the AWS Provider,
v4.9.0 and later retain the same configuration parameters of the `aws_s3_bucket` resource as in v3.x and functionality of the `aws_s3_bucket` resource only differs from v3.x
in that Terraform will only perform drift detection for each of the following parameters if a configuration value is provided:

* `acceleration_status`
* `acl`
* `cors_rule`
* `grant`
* `lifecycle_rule`
* `logging`
* `object_lock_configuration`
* `policy`
* `replication_configuration`
* `request_payer`
* `server_side_encryption_configuration`
* `versioning`
* `website`

Thus, if one of these parameters was once configured and then is entirely removed from an `aws_s3_bucket` resource configuration,
Terraform will not pick up on these changes on a subsequent `terraform plan` or `terraform apply`.

For example, given the following configuration with a single `cors_rule`:

```terraform
resource "aws_s3_bucket" "example" {
  bucket = "yournamehere"

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

When updated to the following configuration without a `cors_rule`:

```terraform
resource "aws_s3_bucket" "example" {
  bucket = "yournamehere"

  # ... other configuration ...
}
```

Terraform CLI with v4.9.0 of the AWS Provider will report back:

```shell
aws_s3_bucket.example: Refreshing state... [id=yournamehere]
...
No changes. Your infrastructure matches the configuration.
```

With that said, to manage changes to these parameters in the `aws_s3_bucket` resource, practitioners should configure each parameter's respective standalone resource
and perform updates directly on those new configurations. The parameters are mapped to the standalone resources as follows:

| `aws_s3_bucket` Parameter              | Standalone Resource                                  |
|----------------------------------------|------------------------------------------------------|
| `acceleration_status`                  | `aws_s3_bucket_accelerate_configuration`             |
| `acl`                                  | `aws_s3_bucket_acl`                                  |
| `cors_rule`                            | `aws_s3_bucket_cors_configuration`                   |
| `grant`                                | `aws_s3_bucket_acl`                                  |
| `lifecycle_rule`                       | `aws_s3_bucket_lifecycle_configuration`              |
| `logging`                              | `aws_s3_bucket_logging`                              |
| `object_lock_configuration`            | `aws_s3_bucket_object_lock_configuration`            |
| `policy`                               | `aws_s3_bucket_policy`                               |
| `replication_configuration`            | `aws_s3_bucket_replication_configuration`            |
| `request_payer`                        | `aws_s3_bucket_request_payment_configuration`        |
| `server_side_encryption_configuration` | `aws_s3_bucket_server_side_encryption_configuration` |
| `versioning`                           | `aws_s3_bucket_versioning`                           |
| `website`                              | `aws_s3_bucket_website_configuration`                |

Going back to the earlier example, given the following configuration:

```terraform
resource "aws_s3_bucket" "example" {
  bucket = "yournamehere"

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

Practitioners can upgrade to v4.9.0 and then introduce the standalone `aws_s3_bucket_cors_configuration` resource, e.g.

```terraform
resource "aws_s3_bucket" "example" {
  bucket = "yournamehere"
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

Depending on the tools available to you, the above configuration can either be directly applied with Terraform or the standalone resource
can be imported into Terraform state. Please refer to each standalone resource's _Import_ documentation for the proper syntax.

Once the standalone resources are managed by Terraform, updates and removal can be performed as needed.

The following sections depict standalone resource adoption per individual parameter. Standalone resource adoption is not required to upgrade but is recommended to ensure drift is detected by Terraform.
The examples below are by no means exhaustive. The aim is to provide important concepts when migrating to a standalone resource whose parameters may not entirely align with the corresponding parameter in the `aws_s3_bucket` resource.

### Migrating to `aws_s3_bucket_accelerate_configuration`

Given this previous configuration:

```terraform
resource "aws_s3_bucket" "example" {
  bucket = "yournamehere"

  # ... other configuration ...
  acceleration_status = "Enabled"
}
```

Update the configuration to:

```terraform
resource "aws_s3_bucket" "example" {
  bucket = "yournamehere"

  # ... other configuration ...
}

resource "aws_s3_bucket_accelerate_configuration" "example" {
  bucket = aws_s3_bucket.example.id
  status = "Enabled"
}
```

### Migrating to `aws_s3_bucket_acl`

#### With `acl`

Given this previous configuration:

```terraform
resource "aws_s3_bucket" "example" {
  bucket = "yournamehere"
  acl    = "private"

  # ... other configuration ...
}
```

Update the configuration to:

```terraform
resource "aws_s3_bucket" "example" {
  bucket = "yournamehere"
}

resource "aws_s3_bucket_acl" "example" {
  bucket = aws_s3_bucket.example.id
  acl    = "private"
}
```

#### With `grant`

Given this previous configuration:

```terraform
resource "aws_s3_bucket" "example" {
  bucket = "yournamehere"

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

Update the configuration to:

```terraform
resource "aws_s3_bucket" "example" {
  bucket = "yournamehere"

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

### Migrating to `aws_s3_bucket_cors_configuration`

Given this previous configuration:

```terraform
resource "aws_s3_bucket" "example" {
  bucket = "yournamehere"

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

Update the configuration to:

```terraform
resource "aws_s3_bucket" "example" {
  bucket = "yournamehere"

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

### Migrating to `aws_s3_bucket_lifecycle_configuration`

~> **Note:** In version `3.x` of the provider, the `lifecycle_rule.id` argument was optional, while in version `4.x`, the `aws_s3_bucket_lifecycle_configuration.rule.id` argument required. Use the AWS CLI s3api [get-bucket-lifecycle-configuration](https://awscli.amazonaws.com/v2/documentation/api/latest/reference/s3api/get-bucket-lifecycle-configuration.html) to get the source bucket's lifecycle configuration to determine the ID.

#### For Lifecycle Rules with no `prefix` previously configured

~> **Note:** When configuring the `rule.filter` configuration block in the new `aws_s3_bucket_lifecycle_configuration` resource, use the AWS CLI s3api [get-bucket-lifecycle-configuration](https://awscli.amazonaws.com/v2/documentation/api/latest/reference/s3api/get-bucket-lifecycle-configuration.html)
to get the source bucket's lifecycle configuration and determine if the `Filter` is configured as `"Filter" : {}` or `"Filter" : { "Prefix": "" }`.
If AWS returns the former, configure `rule.filter` as `filter {}`. Otherwise, neither a `rule.filter` nor `rule.prefix` parameter should be configured as shown here:

Given this previous configuration:

```terraform
resource "aws_s3_bucket" "example" {
  bucket = "yournamehere"

  lifecycle_rule {
    id      = "Keep previous version 30 days, then in Glacier another 60"
    enabled = true

    noncurrent_version_transition {
      days          = 30
      storage_class = "GLACIER"
    }

    noncurrent_version_expiration {
      days = 90
    }
  }

  lifecycle_rule {
    id                                     = "Delete old incomplete multi-part uploads"
    enabled                                = true
    abort_incomplete_multipart_upload_days = 7
  }
}
```

Update the configuration to:

```terraform
resource "aws_s3_bucket" "example" {
  bucket = "yournamehere"

  # ... other configuration ...
}

resource "aws_s3_bucket_lifecycle_configuration" "example" {
  bucket = aws_s3_bucket.example.id

  rule {
    id     = "Keep previous version 30 days, then in Glacier another 60"
    status = "Enabled"

    noncurrent_version_transition {
      noncurrent_days = 30
      storage_class   = "GLACIER"
    }

    noncurrent_version_expiration {
      noncurrent_days = 90
    }
  }

  rule {
    id     = "Delete old incomplete multi-part uploads"
    status = "Enabled"

    abort_incomplete_multipart_upload {
      days_after_initiation = 7
    }
  }
}
```

#### For Lifecycle Rules with `prefix` previously configured as an empty string

Given this previous configuration:

```terraform
resource "aws_s3_bucket" "example" {
  bucket = "yournamehere"

  lifecycle_rule {
    id      = "log-expiration"
    enabled = true
    prefix  = ""

    transition {
      days          = 30
      storage_class = "STANDARD_IA"
    }

    transition {
      days          = 180
      storage_class = "GLACIER"
    }
  }
}
```

Update the configuration to:

```terraform
resource "aws_s3_bucket" "example" {
  bucket = "yournamehere"

  # ... other configuration ...
}

resource "aws_s3_bucket_lifecycle_configuration" "example" {
  bucket = aws_s3_bucket.example.id

  rule {
    id     = "log-expiration"
    status = "Enabled"

    transition {
      days          = 30
      storage_class = "STANDARD_IA"
    }

    transition {
      days          = 180
      storage_class = "GLACIER"
    }
  }
}
```

#### For Lifecycle Rules with `prefix`

Given this previous configuration:

```terraform
resource "aws_s3_bucket" "example" {
  bucket = "yournamehere"

  lifecycle_rule {
    id      = "log-expiration"
    enabled = true
    prefix  = "foobar"

    transition {
      days          = 30
      storage_class = "STANDARD_IA"
    }

    transition {
      days          = 180
      storage_class = "GLACIER"
    }
  }
}
```

Update the configuration to:

```terraform
resource "aws_s3_bucket" "example" {
  bucket = "yournamehere"

  # ... other configuration ...
}

resource "aws_s3_bucket_lifecycle_configuration" "example" {
  bucket = aws_s3_bucket.example.id

  rule {
    id     = "log-expiration"
    status = "Enabled"

    filter {
      prefix = "foobar"
    }

    transition {
      days          = 30
      storage_class = "STANDARD_IA"
    }

    transition {
      days          = 180
      storage_class = "GLACIER"
    }
  }
}
```

#### For Lifecycle Rules with `prefix` and `tags`

Given this previous configuration:

```terraform
resource "aws_s3_bucket" "example" {
  bucket = "yournamehere"

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

Update the configuration to:

```terraform
resource "aws_s3_bucket" "example" {
  bucket = "yournamehere"

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

### Migrating to `aws_s3_bucket_logging`

Given this previous configuration:

```terraform
resource "aws_s3_bucket" "log_bucket" {
  # ... other configuration ...
  bucket = "example-log-bucket"
}

resource "aws_s3_bucket" "example" {
  bucket = "yournamehere"

  # ... other configuration ...
  logging {
    target_bucket = aws_s3_bucket.log_bucket.id
    target_prefix = "log/"
  }
}
```

Update the configuration to:

```terraform
resource "aws_s3_bucket" "log_bucket" {
  bucket = "example-log-bucket"

  # ... other configuration ...
}

resource "aws_s3_bucket" "example" {
  bucket = "yournamehere"

  # ... other configuration ...
}

resource "aws_s3_bucket_logging" "example" {
  bucket        = aws_s3_bucket.example.id
  target_bucket = aws_s3_bucket.log_bucket.id
  target_prefix = "log/"
}
```

### Migrating to `aws_s3_bucket_object_lock_configuration`

Given this previous configuration:

```terraform
resource "aws_s3_bucket" "example" {
  bucket = "yournamehere"

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

Update the configuration to:

```terraform
resource "aws_s3_bucket" "example" {
  bucket = "yournamehere"

  # ... other configuration ...
  object_lock_enabled = true
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

### Migrating to `aws_s3_bucket_policy`

Given this previous configuration:

```terraform
resource "aws_s3_bucket" "example" {
  bucket = "yournamehere"

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
      "Resource": "arn:${data.aws_partition.current.partition}:s3:::yournamehere/*",
      "Sid": "Stmt1446575236270"
    }
  ],
  "Version": "2012-10-17"
}
EOF
}
```

Update the configuration to:

```terraform
resource "aws_s3_bucket" "example" {
  bucket = "yournamehere"

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
      "Resource": "${aws_s3_bucket.example.arn}/*",
      "Sid": "Stmt1446575236270"
    }
  ],
  "Version": "2012-10-17"
}
EOF
}
```

### Migrating to `aws_s3_bucket_replication_configuration`

Given this previous configuration:

```terraform
resource "aws_s3_bucket" "example" {
  provider = aws.central
  bucket   = "yournamehere"

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

Update the configuration to:

```terraform
resource "aws_s3_bucket" "example" {
  provider = aws.central
  bucket   = "yournamehere"

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

### Migrating to `aws_s3_bucket_request_payment_configuration`

Given this previous configuration:

```terraform
resource "aws_s3_bucket" "example" {
  bucket = "yournamehere"

  # ... other configuration ...
  request_payer = "Requester"
}
```

Update the configuration to:

```terraform
resource "aws_s3_bucket" "example" {
  bucket = "yournamehere"

  # ... other configuration ...
}

resource "aws_s3_bucket_request_payment_configuration" "example" {
  bucket = aws_s3_bucket.example.id
  payer  = "Requester"
}
```

### Migrating to `aws_s3_bucket_server_side_encryption_configuration`

Given this previous configuration:

```terraform
resource "aws_s3_bucket" "example" {
  bucket = "yournamehere"

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

Update the configuration to:

```terraform
resource "aws_s3_bucket" "example" {
  bucket = "yournamehere"

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

### Migrating to `aws_s3_bucket_versioning`

~> **NOTE:** As `aws_s3_bucket_versioning` is a separate resource, any S3 objects for which versioning is important (_e.g._, a truststore for mutual TLS authentication) must implicitly or explicitly depend on the `aws_s3_bucket_versioning` resource. Otherwise, the S3 objects may be created before versioning has been set. [See below](#ensure-objects-depend-on-versioning) for an example. Also note that AWS recommends waiting 15 minutes after enabling versioning on a bucket before putting or deleting objects in/from the bucket.

#### Buckets With Versioning Enabled

Given this previous configuration:

```terraform
resource "aws_s3_bucket" "example" {
  bucket = "yournamehere"

  # ... other configuration ...
  versioning {
    enabled = true
  }
}
```

Update the configuration to:

```terraform
resource "aws_s3_bucket" "example" {
  bucket = "yournamehere"

  # ... other configuration ...
}

resource "aws_s3_bucket_versioning" "example" {
  bucket = aws_s3_bucket.example.id
  versioning_configuration {
    status = "Enabled"
  }
}
```

#### Buckets With Versioning Disabled or Suspended

Depending on the version of the Terraform AWS Provider you are migrating from, the interpretation of `versioning.enabled = false`
in your `aws_s3_bucket` resource will differ and thus the migration to the `aws_s3_bucket_versioning` resource will also differ as follows.

If you are migrating from the Terraform AWS Provider `v3.70.0` or later:

* For new S3 buckets, `enabled = false` is synonymous to `Disabled`.
* For existing S3 buckets, `enabled = false` is synonymous to `Suspended`.

If you are migrating from an earlier version of the Terraform AWS Provider:

* For both new and existing S3 buckets, `enabled = false` is synonymous to `Suspended`.

Given this previous configuration:

```terraform
resource "aws_s3_bucket" "example" {
  bucket = "yournamehere"

  # ... other configuration ...
  versioning {
    enabled = false
  }
}
```

Update the configuration to one of the following:

* If migrating from Terraform AWS Provider `v3.70.0` or later and bucket versioning was never enabled:

  ```terraform
  resource "aws_s3_bucket" "example" {
    bucket = "yournamehere"
  
    # ... other configuration ...
  }
  
  resource "aws_s3_bucket_versioning" "example" {
    bucket = aws_s3_bucket.example.id
    versioning_configuration {
      status = "Disabled"
    }
  }
  ```

* If migrating from Terraform AWS Provider `v3.70.0` or later and bucket versioning was enabled at one point:

  ```terraform
  resource "aws_s3_bucket" "example" {
    bucket = "yournamehere"
  
    # ... other configuration ...
  }
  
  resource "aws_s3_bucket_versioning" "example" {
    bucket = aws_s3_bucket.example.id
    versioning_configuration {
      status = "Suspended"
    }
  }
  ```

* If migrating from an earlier version of Terraform AWS Provider:

  ```terraform
  resource "aws_s3_bucket" "example" {
    bucket = "yournamehere"
  
    # ... other configuration ...
  }
  
  resource "aws_s3_bucket_versioning" "example" {
    bucket = aws_s3_bucket.example.id
    versioning_configuration {
      status = "Suspended"
    }
  }
  ```

#### Ensure Objects Depend on Versioning

When you create an object whose `version_id` you need and an `aws_s3_bucket_versioning` resource in the same configuration, you are more likely to have success by ensuring the `s3_object` depends either implicitly (see below) or explicitly (i.e., using `depends_on = [aws_s3_bucket_versioning.example]`) on the `aws_s3_bucket_versioning` resource.

~> **NOTE:** For critical and/or production S3 objects, do not create a bucket, enable versioning, and create an object in the bucket within the same configuration. Doing so will not allow the AWS-recommended 15 minutes between enabling versioning and writing to the bucket.

This example shows the `aws_s3_object.example` depending implicitly on the versioning resource through the reference to `aws_s3_bucket_versioning.example.bucket` to define `bucket`:

```terraform
resource "aws_s3_bucket" "example" {
  bucket = "yotto"
}

resource "aws_s3_bucket_versioning" "example" {
  bucket = aws_s3_bucket.example.id

  versioning_configuration {
    status = "Enabled"
  }
}

resource "aws_s3_object" "example" {
  bucket = aws_s3_bucket_versioning.example.bucket
  key    = "droeloe"
  source = "example.txt"
}
```

### Migrating to `aws_s3_bucket_website_configuration`

Given this previous configuration:

```terraform
resource "aws_s3_bucket" "example" {
  bucket = "yournamehere"

  # ... other configuration ...
  website {
    index_document = "index.html"
    error_document = "error.html"
  }
}
```

Update the configuration to:

```terraform
resource "aws_s3_bucket" "example" {
  bucket = "yournamehere"

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

Given this previous configuration that uses the `aws_s3_bucket` parameter `website_domain` with `aws_route53_record`:

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

Update the configuration to use the `aws_s3_bucket_website_configuration` resource and its `website_domain` parameter:

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

## S3 Bucket Refactor

~> **NOTE:** This only applies to v4.0.0 through v4.8.0 of the AWS Provider, which introduce significant breaking
changes to the `aws_s3_bucket` resource. We recommend upgrading to v4.9.0 of the AWS Provider instead. See the section above, [Changes to S3 Bucket Drift Detection](#changes-to-s3-bucket-drift-detection), for additional upgrade considerations.

To help distribute the management of S3 bucket settings via independent resources, various arguments and attributes in the `aws_s3_bucket` resource have become **read-only**.

Configurations dependent on these arguments should be updated to use the corresponding `aws_s3_bucket_*` resource in order to prevent Terraform from reporting “unconfigurable attribute” errors for read-only arguments. Once updated, it is recommended to import new `aws_s3_bucket_*` resources into Terraform state.

In the event practitioners do not anticipate future modifications to the S3 bucket settings associated with these read-only arguments or drift detection is not needed, these read-only arguments should be removed from `aws_s3_bucket` resource configurations in order to prevent Terraform from reporting “unconfigurable attribute” errors; the states of these arguments will be preserved but are subject to change with modifications made outside Terraform.

~> **NOTE:** Each of the new `aws_s3_bucket_*` resources relies on S3 API calls that utilize a `PUT` action in order to modify the target S3 bucket. These calls follow standard HTTP methods for REST APIs, and therefore **should** handle situations where the target configuration already exists. While it is not strictly necessary to import new `aws_s3_bucket_*` resources where the updated configuration matches the configuration used in previous versions of the AWS provider, skipping this step will lead to a diff in the first plan after a configuration change indicating that any new `aws_s3_bucket_*` resources will be created, making it more difficult to determine whether the appropriate actions will be taken.

### `acceleration_status` Argument

Switch your Terraform configuration to the [`aws_s3_bucket_accelerate_configuration` resource](/docs/providers/aws/r/s3_bucket_accelerate_configuration.html) instead.

For example, given this previous configuration:

```terraform
resource "aws_s3_bucket" "example" {
  bucket = "yournamehere"

  # ... other configuration ...
  acceleration_status = "Enabled"
}
```

You will get the following error after upgrading:

```
│ Error: Value for unconfigurable attribute
│
│   with aws_s3_bucket.example,
│   on main.tf line 1, in resource "aws_s3_bucket" "example":
│    1: resource "aws_s3_bucket" "example" {
│
│ Can't configure a value for "acceleration_status": its value will be decided automatically based on the result of applying this configuration.
```

Since `acceleration_status` is now read only, update your configuration to use the `aws_s3_bucket_accelerate_configuration`
resource and remove `acceleration_status` in the `aws_s3_bucket` resource:

```terraform
resource "aws_s3_bucket" "example" {
  bucket = "yournamehere"

  # ... other configuration ...
}

resource "aws_s3_bucket_accelerate_configuration" "example" {
  bucket = aws_s3_bucket.example.id
  status = "Enabled"
}
```

Run `terraform import` on each new resource, _e.g._,

```shell
$ terraform import aws_s3_bucket_accelerate_configuration.example yournamehere
aws_s3_bucket_accelerate_configuration.example: Importing from ID "yournamehere"...
aws_s3_bucket_accelerate_configuration.example: Import prepared!
  Prepared aws_s3_bucket_accelerate_configuration for import
aws_s3_bucket_accelerate_configuration.example: Refreshing state... [id=yournamehere]

Import successful!

The resources that were imported are shown above. These resources are now in
your Terraform state and will henceforth be managed by Terraform.
```

### `acl` Argument

Switch your Terraform configuration to the [`aws_s3_bucket_acl` resource](/docs/providers/aws/r/s3_bucket_acl.html) instead.

For example, given this previous configuration:

```terraform
resource "aws_s3_bucket" "example" {
  bucket = "yournamehere"
  acl    = "private"

  # ... other configuration ...
}
```

You will get the following error after upgrading:

```
│ Error: Value for unconfigurable attribute
│
│   with aws_s3_bucket.example,
│   on main.tf line 1, in resource "aws_s3_bucket" "example":
│    1: resource "aws_s3_bucket" "example" {
│
│ Can't configure a value for "acl": its value will be decided automatically based on the result of applying this configuration.
```

Since `acl` is now read only, update your configuration to use the `aws_s3_bucket_acl`
resource and remove the `acl` argument in the `aws_s3_bucket` resource:

```terraform
resource "aws_s3_bucket" "example" {
  bucket = "yournamehere"
}

resource "aws_s3_bucket_acl" "example" {
  bucket = aws_s3_bucket.example.id
  acl    = "private"
}
```

~> **NOTE:** When importing into `aws_s3_bucket_acl`, make sure you use the S3 bucket name (_e.g._, `yournamehere` in the example above) as part of the ID, and _not_ the Terraform bucket configuration name (_e.g._, `example` in the example above).

Run `terraform import` on each new resource, _e.g._,

```shell
$ terraform import aws_s3_bucket_acl.example yournamehere,private
aws_s3_bucket_acl.example: Importing from ID "yournamehere,private"...
aws_s3_bucket_acl.example: Import prepared!
  Prepared aws_s3_bucket_acl for import
aws_s3_bucket_acl.example: Refreshing state... [id=yournamehere,private]

Import successful!

The resources that were imported are shown above. These resources are now in
your Terraform state and will henceforth be managed by Terraform.
```

### `cors_rule` Argument

Switch your Terraform configuration to the [`aws_s3_bucket_cors_configuration` resource](/docs/providers/aws/r/s3_bucket_cors_configuration.html) instead.

For example, given this previous configuration:

```terraform
resource "aws_s3_bucket" "example" {
  bucket = "yournamehere"

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

You will get the following error after upgrading:

```
│ Error: Value for unconfigurable attribute
│
│   with aws_s3_bucket.example,
│   on main.tf line 1, in resource "aws_s3_bucket" "example":
│    1: resource "aws_s3_bucket" "example" {
│
│ Can't configure a value for "cors_rule": its value will be decided automatically based on the result of applying this configuration.
```

Since `cors_rule` is now read only, update your configuration to use the `aws_s3_bucket_cors_configuration`
resource and remove `cors_rule` and its nested arguments in the `aws_s3_bucket` resource:

```terraform
resource "aws_s3_bucket" "example" {
  bucket = "yournamehere"

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

Run `terraform import` on each new resource, _e.g._,

```shell
$ terraform import aws_s3_bucket_cors_configuration.example yournamehere
aws_s3_bucket_cors_configuration.example: Importing from ID "yournamehere"...
aws_s3_bucket_cors_configuration.example: Import prepared!
  Prepared aws_s3_bucket_cors_configuration for import
aws_s3_bucket_cors_configuration.example: Refreshing state... [id=yournamehere]

Import successful!

The resources that were imported are shown above. These resources are now in
your Terraform state and will henceforth be managed by Terraform.
```

### `grant` Argument

Switch your Terraform configuration to the [`aws_s3_bucket_acl` resource](/docs/providers/aws/r/s3_bucket_acl.html) instead.

For example, given this previous configuration:

```terraform
resource "aws_s3_bucket" "example" {
  bucket = "yournamehere"

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

You will get the following error after upgrading:

```
│ Error: Value for unconfigurable attribute
│
│   with aws_s3_bucket.example,
│   on main.tf line 1, in resource "aws_s3_bucket" "example":
│    1: resource "aws_s3_bucket" "example" {
│
│ Can't configure a value for "grant": its value will be decided automatically based on the result of applying this configuration.
```

Since `grant` is now read only, update your configuration to use the `aws_s3_bucket_acl`
resource and remove `grant` in the `aws_s3_bucket` resource:

```terraform
resource "aws_s3_bucket" "example" {
  bucket = "yournamehere"

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

Run `terraform import` on each new resource, _e.g._,

```shell
$ terraform import aws_s3_bucket_acl.example yournamehere
aws_s3_bucket_acl.example: Importing from ID "yournamehere"...
aws_s3_bucket_acl.example: Import prepared!
  Prepared aws_s3_bucket_acl for import
aws_s3_bucket_acl.example: Refreshing state... [id=yournamehere]

Import successful!

The resources that were imported are shown above. These resources are now in
your Terraform state and will henceforth be managed by Terraform.
```

### `lifecycle_rule` Argument

Switch your Terraform configuration to the [`aws_s3_bucket_lifecycle_configuration` resource](/docs/providers/aws/r/s3_bucket_lifecycle_configuration.html) instead.

#### For Lifecycle Rules with no `prefix` previously configured

For example, given this previous configuration:

```terraform
resource "aws_s3_bucket" "example" {
  bucket = "yournamehere"

  lifecycle_rule {
    id      = "Keep previous version 30 days, then in Glacier another 60"
    enabled = true

    noncurrent_version_transition {
      days          = 30
      storage_class = "GLACIER"
    }

    noncurrent_version_expiration {
      days = 90
    }
  }

  lifecycle_rule {
    id                                     = "Delete old incomplete multi-part uploads"
    enabled                                = true
    abort_incomplete_multipart_upload_days = 7
  }
}
```

You will receive the following error after upgrading:

```
│ Error: Value for unconfigurable attribute
│
│   with aws_s3_bucket.example,
│   on main.tf line 1, in resource "aws_s3_bucket" "example":
│    1: resource "aws_s3_bucket" "example" {
│
│ Can't configure a value for "lifecycle_rule": its value will be decided automatically based on the result of applying this configuration.
```

Since the `lifecycle_rule` argument changed to read-only, update the configuration to use the `aws_s3_bucket_lifecycle_configuration`
resource and remove `lifecycle_rule` and its nested arguments in the `aws_s3_bucket` resource.

~> **Note:** When configuring the `rule.filter` configuration block in the new `aws_s3_bucket_lifecycle_configuration` resource, use the AWS CLI s3api [get-bucket-lifecycle-configuration](https://awscli.amazonaws.com/v2/documentation/api/latest/reference/s3api/get-bucket-lifecycle-configuration.html)
to get the source bucket's lifecycle configuration and determine if the `Filter` is configured as `"Filter" : {}` or `"Filter" : { "Prefix": "" }`.
If AWS returns the former, configure `rule.filter` as `filter {}`. Otherwise, neither a `rule.filter` nor `rule.prefix` parameter should be configured as shown here:

```terraform
resource "aws_s3_bucket" "example" {
  bucket = "yournamehere"

  # ... other configuration ...
}

resource "aws_s3_bucket_lifecycle_configuration" "example" {
  bucket = aws_s3_bucket.example.id

  rule {
    id     = "Keep previous version 30 days, then in Glacier another 60"
    status = "Enabled"

    noncurrent_version_transition {
      noncurrent_days = 30
      storage_class   = "GLACIER"
    }

    noncurrent_version_expiration {
      noncurrent_days = 90
    }
  }

  rule {
    id     = "Delete old incomplete multi-part uploads"
    status = "Enabled"

    abort_incomplete_multipart_upload {
      days_after_initiation = 7
    }
  }
}
```

Run `terraform import` on each new resource, _e.g._,

```shell
$ terraform import aws_s3_bucket_lifecycle_configuration.example yournamehere
aws_s3_bucket_lifecycle_configuration.example: Importing from ID "yournamehere"...
aws_s3_bucket_lifecycle_configuration.example: Import prepared!
  Prepared aws_s3_bucket_lifecycle_configuration for import
aws_s3_bucket_lifecycle_configuration.example: Refreshing state... [id=yournamehere]

Import successful!

The resources that were imported are shown above. These resources are now in
your Terraform state and will henceforth be managed by Terraform.
```

#### For Lifecycle Rules with `prefix` previously configured as an empty string

For example, given this configuration:

```terraform
resource "aws_s3_bucket" "example" {
  bucket = "yournamehere"

  lifecycle_rule {
    id      = "log-expiration"
    enabled = true
    prefix  = ""

    transition {
      days          = 30
      storage_class = "STANDARD_IA"
    }

    transition {
      days          = 180
      storage_class = "GLACIER"
    }
  }
}
```

You will receive the following error after upgrading:

```
│ Error: Value for unconfigurable attribute
│
│   with aws_s3_bucket.example,
│   on main.tf line 1, in resource "aws_s3_bucket" "example":
│    1: resource "aws_s3_bucket" "example" {
│
│ Can't configure a value for "lifecycle_rule": its value will be decided automatically based on the result of applying this configuration.
```

Since the `lifecycle_rule` argument changed to read-only, update the configuration to use the `aws_s3_bucket_lifecycle_configuration`
resource and remove `lifecycle_rule` and its nested arguments in the `aws_s3_bucket` resource:

```terraform
resource "aws_s3_bucket" "example" {
  bucket = "yournamehere"

  # ... other configuration ...
}

resource "aws_s3_bucket_lifecycle_configuration" "example" {
  bucket = aws_s3_bucket.example.id

  rule {
    id     = "log-expiration"
    status = "Enabled"

    transition {
      days          = 30
      storage_class = "STANDARD_IA"
    }

    transition {
      days          = 180
      storage_class = "GLACIER"
    }
  }
}
```

Run `terraform import` on each new resource, _e.g._,

```shell
$ terraform import aws_s3_bucket_lifecycle_configuration.example yournamehere
aws_s3_bucket_lifecycle_configuration.example: Importing from ID "yournamehere"...
aws_s3_bucket_lifecycle_configuration.example: Import prepared!
  Prepared aws_s3_bucket_lifecycle_configuration for import
aws_s3_bucket_lifecycle_configuration.example: Refreshing state... [id=yournamehere]

Import successful!

The resources that were imported are shown above. These resources are now in
your Terraform state and will henceforth be managed by Terraform.
```

#### For Lifecycle Rules with `prefix`

For example, given this configuration:

```terraform
resource "aws_s3_bucket" "example" {
  bucket = "yournamehere"

  lifecycle_rule {
    id      = "log-expiration"
    enabled = true
    prefix  = "foobar"

    transition {
      days          = 30
      storage_class = "STANDARD_IA"
    }

    transition {
      days          = 180
      storage_class = "GLACIER"
    }
  }
}
```

You will receive the following error after upgrading:

```
│ Error: Value for unconfigurable attribute
│
│   with aws_s3_bucket.example,
│   on main.tf line 1, in resource "aws_s3_bucket" "example":
│    1: resource "aws_s3_bucket" "example" {
│
│ Can't configure a value for "lifecycle_rule": its value will be decided automatically based on the result of applying this configuration.
```

Since the `lifecycle_rule` argument changed to read-only, update the configuration to use the `aws_s3_bucket_lifecycle_configuration`
resource and remove `lifecycle_rule` and its nested arguments in the `aws_s3_bucket` resource:

```terraform
resource "aws_s3_bucket" "example" {
  bucket = "yournamehere"

  # ... other configuration ...
}

resource "aws_s3_bucket_lifecycle_configuration" "example" {
  bucket = aws_s3_bucket.example.id

  rule {
    id     = "log-expiration"
    status = "Enabled"

    filter {
      prefix = "foobar"
    }

    transition {
      days          = 30
      storage_class = "STANDARD_IA"
    }

    transition {
      days          = 180
      storage_class = "GLACIER"
    }
  }
}
```

Run `terraform import` on each new resource, _e.g._,

```shell
$ terraform import aws_s3_bucket_lifecycle_configuration.example yournamehere
aws_s3_bucket_lifecycle_configuration.example: Importing from ID "yournamehere"...
aws_s3_bucket_lifecycle_configuration.example: Import prepared!
  Prepared aws_s3_bucket_lifecycle_configuration for import
aws_s3_bucket_lifecycle_configuration.example: Refreshing state... [id=yournamehere]

Import successful!

The resources that were imported are shown above. These resources are now in
your Terraform state and will henceforth be managed by Terraform.
```

#### For Lifecycle Rules with `prefix` and `tags`

For example, given this previous configuration:

```terraform
resource "aws_s3_bucket" "example" {
  bucket = "yournamehere"

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

You will get the following error after upgrading:

```
│ Error: Value for unconfigurable attribute
│
│   with aws_s3_bucket.example,
│   on main.tf line 1, in resource "aws_s3_bucket" "example":
│    1: resource "aws_s3_bucket" "example" {
│
│ Can't configure a value for "lifecycle_rule": its value will be decided automatically based on the result of applying this configuration.
```

Since `lifecycle_rule` is now read only, update your configuration to use the `aws_s3_bucket_lifecycle_configuration`
resource and remove `lifecycle_rule` and its nested arguments in the `aws_s3_bucket` resource:

```terraform
resource "aws_s3_bucket" "example" {
  bucket = "yournamehere"

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

Run `terraform import` on each new resource, _e.g._,

```shell
$ terraform import aws_s3_bucket_lifecycle_configuration.example yournamehere
aws_s3_bucket_lifecycle_configuration.example: Importing from ID "yournamehere"...
aws_s3_bucket_lifecycle_configuration.example: Import prepared!
  Prepared aws_s3_bucket_lifecycle_configuration for import
aws_s3_bucket_lifecycle_configuration.example: Refreshing state... [id=yournamehere]

Import successful!

The resources that were imported are shown above. These resources are now in
your Terraform state and will henceforth be managed by Terraform.
```

### `logging` Argument

Switch your Terraform configuration to the [`aws_s3_bucket_logging` resource](/docs/providers/aws/r/s3_bucket_logging.html) instead.

For example, given this previous configuration:

```terraform
resource "aws_s3_bucket" "log_bucket" {
  # ... other configuration ...
  bucket = "example-log-bucket"
}

resource "aws_s3_bucket" "example" {
  bucket = "yournamehere"

  # ... other configuration ...
  logging {
    target_bucket = aws_s3_bucket.log_bucket.id
    target_prefix = "log/"
  }
}
```

You will get the following error after upgrading:

```
│ Error: Value for unconfigurable attribute
│
│   with aws_s3_bucket.example,
│   on main.tf line 1, in resource "aws_s3_bucket" "example":
│    1: resource "aws_s3_bucket" "example" {
│
│ Can't configure a value for "logging": its value will be decided automatically based on the result of applying this configuration.
```

Since `logging` is now read only, update your configuration to use the `aws_s3_bucket_logging`
resource and remove `logging` and its nested arguments in the `aws_s3_bucket` resource:

```terraform
resource "aws_s3_bucket" "log_bucket" {
  bucket = "example-log-bucket"

  # ... other configuration ...
}

resource "aws_s3_bucket" "example" {
  bucket = "yournamehere"

  # ... other configuration ...
}

resource "aws_s3_bucket_logging" "example" {
  bucket        = aws_s3_bucket.example.id
  target_bucket = aws_s3_bucket.log_bucket.id
  target_prefix = "log/"
}
```

Run `terraform import` on each new resource, _e.g._,

```shell
$ terraform import aws_s3_bucket_logging.example yournamehere
aws_s3_bucket_logging.example: Importing from ID "yournamehere"...
aws_s3_bucket_logging.example: Import prepared!
  Prepared aws_s3_bucket_logging for import
aws_s3_bucket_logging.example: Refreshing state... [id=yournamehere]

Import successful!

The resources that were imported are shown above. These resources are now in
your Terraform state and will henceforth be managed by Terraform.
```

### `object_lock_configuration` `rule` Argument

Switch your Terraform configuration to the [`aws_s3_bucket_object_lock_configuration` resource](/docs/providers/aws/r/s3_bucket_object_lock_configuration.html) instead.

For example, given this previous configuration:

```terraform
resource "aws_s3_bucket" "example" {
  bucket = "yournamehere"

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

You will get the following error after upgrading:

```
│ Error: Value for unconfigurable attribute
│
│   with aws_s3_bucket.example,
│   on main.tf line 1, in resource "aws_s3_bucket" "example":
│    1: resource "aws_s3_bucket" "example" {
│
│ Can't configure a value for "object_lock_configuration.0.rule": its value will be decided automatically based on the result of applying this configuration.
```

Since the `rule` argument of the `object_lock_configuration` configuration block changed to read-only, update your configuration to use the `aws_s3_bucket_object_lock_configuration`
resource and remove `rule` and its nested arguments in the `aws_s3_bucket` resource:

```terraform
resource "aws_s3_bucket" "example" {
  bucket = "yournamehere"

  # ... other configuration ...
  object_lock_enabled = true
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

Run `terraform import` on each new resource, _e.g._,

```shell
$ terraform import aws_s3_bucket_object_lock_configuration.example yournamehere
aws_s3_bucket_object_lock_configuration.example: Importing from ID "yournamehere"...
aws_s3_bucket_object_lock_configuration.example: Import prepared!
  Prepared aws_s3_bucket_object_lock_configuration for import
aws_s3_bucket_object_lock_configuration.example: Refreshing state... [id=yournamehere]

Import successful!

The resources that were imported are shown above. These resources are now in
your Terraform state and will henceforth be managed by Terraform.
```

### `policy` Argument

Switch your Terraform configuration to the [`aws_s3_bucket_policy` resource](/docs/providers/aws/r/s3_bucket_policy.html) instead.

For example, given this previous configuration:

```terraform
resource "aws_s3_bucket" "example" {
  bucket = "yournamehere"

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
      "Resource": "arn:${data.aws_partition.current.partition}:s3:::yournamehere/*",
      "Sid": "Stmt1446575236270"
    }
  ],
  "Version": "2012-10-17"
}
EOF
}
```

You will get the following error after upgrading:

```
│ Error: Value for unconfigurable attribute
│
│   with aws_s3_bucket.accesslogs_bucket,
│   on main.tf line 1, in resource "aws_s3_bucket" "accesslogs_bucket":
│    1: resource "aws_s3_bucket" "accesslogs_bucket" {
│
│ Can't configure a value for "policy": its value will be decided automatically based on the result of applying this configuration.
```

Since `policy` is now read only, update your configuration to use the `aws_s3_bucket_policy`
resource and remove `policy` in the `aws_s3_bucket` resource:

```terraform
resource "aws_s3_bucket" "example" {
  bucket = "yournamehere"

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
      "Resource": "${aws_s3_bucket.example.arn}/*",
      "Sid": "Stmt1446575236270"
    }
  ],
  "Version": "2012-10-17"
}
EOF
}
```

Run `terraform import` on each new resource, _e.g._,

```shell
$ terraform import aws_s3_bucket_policy.example yournamehere
aws_s3_bucket_policy.example: Importing from ID "yournamehere"...
aws_s3_bucket_policy.example: Import prepared!
  Prepared aws_s3_bucket_policy for import
aws_s3_bucket_policy.example: Refreshing state... [id=yournamehere]

Import successful!

The resources that were imported are shown above. These resources are now in
your Terraform state and will henceforth be managed by Terraform.
```

### `replication_configuration` Argument

Switch your Terraform configuration to the [`aws_s3_bucket_replication_configuration` resource](/docs/providers/aws/r/s3_bucket_replication_configuration.html) instead.

For example, given this previous configuration:

```terraform
resource "aws_s3_bucket" "example" {
  provider = aws.central
  bucket   = "yournamehere"

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

You will get the following error after upgrading:

```
│ Error: Value for unconfigurable attribute
│
│   with aws_s3_bucket.source,
│   on main.tf line 1, in resource "aws_s3_bucket" "source":
│    1: resource "aws_s3_bucket" "source" {
│
│ Can't configure a value for "replication_configuration": its value will be decided automatically based on the result of applying this configuration.
```

Since `replication_configuration` is now read only, update your configuration to use the `aws_s3_bucket_replication_configuration`
resource and remove `replication_configuration` and its nested arguments in the `aws_s3_bucket` resource:

```terraform
resource "aws_s3_bucket" "example" {
  provider = aws.central
  bucket   = "yournamehere"

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

Run `terraform import` on each new resource, _e.g._,

```shell
$ terraform import aws_s3_bucket_replication_configuration.example yournamehere
aws_s3_bucket_replication_configuration.example: Importing from ID "yournamehere"...
aws_s3_bucket_replication_configuration.example: Import prepared!
  Prepared aws_s3_bucket_replication_configuration for import
aws_s3_bucket_replication_configuration.example: Refreshing state... [id=yournamehere]

Import successful!

The resources that were imported are shown above. These resources are now in
your Terraform state and will henceforth be managed by Terraform.
```

### `request_payer` Argument

Switch your Terraform configuration to the [`aws_s3_bucket_request_payment_configuration` resource](/docs/providers/aws/r/s3_bucket_request_payment_configuration.html) instead.

For example, given this previous configuration:

```terraform
resource "aws_s3_bucket" "example" {
  bucket = "yournamehere"

  # ... other configuration ...
  request_payer = "Requester"
}
```

You will get the following error after upgrading:

```
│ Error: Value for unconfigurable attribute
│
│   with aws_s3_bucket.example,
│   on main.tf line 1, in resource "aws_s3_bucket" "example":
│    1: resource "aws_s3_bucket" "example" {
│
│ Can't configure a value for "request_payer": its value will be decided automatically based on the result of applying this configuration.
```

Since `request_payer` is now read only, update your configuration to use the `aws_s3_bucket_request_payment_configuration`
resource and remove `request_payer` in the `aws_s3_bucket` resource:

```terraform
resource "aws_s3_bucket" "example" {
  bucket = "yournamehere"

  # ... other configuration ...
}

resource "aws_s3_bucket_request_payment_configuration" "example" {
  bucket = aws_s3_bucket.example.id
  payer  = "Requester"
}
```

Run `terraform import` on each new resource, _e.g._,

```shell
$ terraform import aws_s3_bucket_request_payment_configuration.example yournamehere
aws_s3_bucket_request_payment_configuration.example: Importing from ID "yournamehere"...
aws_s3_bucket_request_payment_configuration.example: Import prepared!
  Prepared aws_s3_bucket_request_payment_configuration for import
aws_s3_bucket_request_payment_configuration.example: Refreshing state... [id=yournamehere]

Import successful!

The resources that were imported are shown above. These resources are now in
your Terraform state and will henceforth be managed by Terraform.
```

### `server_side_encryption_configuration` Argument

Switch your Terraform configuration to the [`aws_s3_bucket_server_side_encryption_configuration` resource](/docs/providers/aws/r/s3_bucket_server_side_encryption_configuration.html) instead.

For example, given this previous configuration:

```terraform
resource "aws_s3_bucket" "example" {
  bucket = "yournamehere"

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

You will get the following error after upgrading:

```
│ Error: Value for unconfigurable attribute
│
│   with aws_s3_bucket.example,
│   on main.tf line 1, in resource "aws_s3_bucket" "example":
│    1: resource "aws_s3_bucket" "example" {
│
│ Can't configure a value for "server_side_encryption_configuration": its value will be decided automatically based on the result of applying this configuration.
```

Since `server_side_encryption_configuration` is now read only, update your configuration to use the `aws_s3_bucket_server_side_encryption_configuration`
resource and remove `server_side_encryption_configuration` and its nested arguments in the `aws_s3_bucket` resource:

```terraform
resource "aws_s3_bucket" "example" {
  bucket = "yournamehere"

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

Run `terraform import` on each new resource, _e.g._,

```shell
$ terraform import aws_s3_bucket_server_side_encryption_configuration.example yournamehere
aws_s3_bucket_server_side_encryption_configuration.example: Importing from ID "yournamehere"...
aws_s3_bucket_server_side_encryption_configuration.example: Import prepared!
  Prepared aws_s3_bucket_server_side_encryption_configuration for import
aws_s3_bucket_server_side_encryption_configuration.example: Refreshing state... [id=yournamehere]

Import successful!

The resources that were imported are shown above. These resources are now in
your Terraform state and will henceforth be managed by Terraform.
```

### `versioning` Argument

Switch your Terraform configuration to the [`aws_s3_bucket_versioning` resource](/docs/providers/aws/r/s3_bucket_versioning.html) instead.

~> **NOTE:** As `aws_s3_bucket_versioning` is a separate resource, any S3 objects for which versioning is important (_e.g._, a truststore for mutual TLS authentication) must implicitly or explicitly depend on the `aws_s3_bucket_versioning` resource. Otherwise, the S3 objects may be created before versioning has been set. [See below](#ensure-objects-depend-on-versioning) for an example. Also note that AWS recommends waiting 15 minutes after enabling versioning on a bucket before putting or deleting objects in/from the bucket.

#### Buckets With Versioning Enabled

Given this previous configuration:

```terraform
resource "aws_s3_bucket" "example" {
  bucket = "yournamehere"

  # ... other configuration ...
  versioning {
    enabled = true
  }
}
```

You will get the following error after upgrading:

```
│ Error: Value for unconfigurable attribute
│
│   with aws_s3_bucket.example,
│   on main.tf line 1, in resource "aws_s3_bucket" "example":
│    1: resource "aws_s3_bucket" "example" {
│
│ Can't configure a value for "versioning": its value will be decided automatically based on the result of applying this configuration.
```

Since `versioning` is now read only, update your configuration to use the `aws_s3_bucket_versioning`
resource and remove `versioning` and its nested arguments in the `aws_s3_bucket` resource:

```terraform
resource "aws_s3_bucket" "example" {
  bucket = "yournamehere"

  # ... other configuration ...
}

resource "aws_s3_bucket_versioning" "example" {
  bucket = aws_s3_bucket.example.id
  versioning_configuration {
    status = "Enabled"
  }
}
```

Run `terraform import` on each new resource, _e.g._,

```shell
$ terraform import aws_s3_bucket_versioning.example yournamehere
aws_s3_bucket_versioning.example: Importing from ID "yournamehere"...
aws_s3_bucket_versioning.example: Import prepared!
  Prepared aws_s3_bucket_versioning for import
aws_s3_bucket_versioning.example: Refreshing state... [id=yournamehere]

Import successful!

The resources that were imported are shown above. These resources are now in
your Terraform state and will henceforth be managed by Terraform.
```

#### Buckets With Versioning Disabled or Suspended

Depending on the version of the Terraform AWS Provider you are migrating from, the interpretation of `versioning.enabled = false`
in your `aws_s3_bucket` resource will differ and thus the migration to the `aws_s3_bucket_versioning` resource will also differ as follows.

If you are migrating from the Terraform AWS Provider `v3.70.0` or later:

* For new S3 buckets, `enabled = false` is synonymous to `Disabled`.
* For existing S3 buckets, `enabled = false` is synonymous to `Suspended`.

If you are migrating from an earlier version of the Terraform AWS Provider:

* For both new and existing S3 buckets, `enabled = false` is synonymous to `Suspended`.

Given this previous configuration :

```terraform
resource "aws_s3_bucket" "example" {
  bucket = "yournamehere"

  # ... other configuration ...
  versioning {
    enabled = false
  }
}
```

You will get the following error after upgrading:

```
│ Error: Value for unconfigurable attribute
│
│   with aws_s3_bucket.example,
│   on main.tf line 1, in resource "aws_s3_bucket" "example":
│    1: resource "aws_s3_bucket" "example" {
│
│ Can't configure a value for "versioning": its value will be decided automatically based on the result of applying this configuration.
```

Since `versioning` is now read only, update your configuration to use the `aws_s3_bucket_versioning`
resource and remove `versioning` and its nested arguments in the `aws_s3_bucket` resource.

* If migrating from Terraform AWS Provider `v3.70.0` or later and bucket versioning was never enabled:

  ```terraform
  resource "aws_s3_bucket" "example" {
    bucket = "yournamehere"
  
    # ... other configuration ...
  }
  
  resource "aws_s3_bucket_versioning" "example" {
    bucket = aws_s3_bucket.example.id
    versioning_configuration {
      status = "Disabled"
    }
  }
  ```

* If migrating from Terraform AWS Provider `v3.70.0` or later and bucket versioning was enabled at one point:

  ```terraform
  resource "aws_s3_bucket" "example" {
    bucket = "yournamehere"
  
    # ... other configuration ...
  }
  
  resource "aws_s3_bucket_versioning" "example" {
    bucket = aws_s3_bucket.example.id
    versioning_configuration {
      status = "Suspended"
    }
  }
  ```

* If migrating from an earlier version of Terraform AWS Provider:

  ```terraform
  resource "aws_s3_bucket" "example" {
    bucket = "yournamehere"
  
    # ... other configuration ...
  }
  
  resource "aws_s3_bucket_versioning" "example" {
    bucket = aws_s3_bucket.example.id
    versioning_configuration {
      status = "Suspended"
    }
  }
  ```

Run `terraform import` on each new resource, _e.g._,

```shell
$ terraform import aws_s3_bucket_versioning.example yournamehere
aws_s3_bucket_versioning.example: Importing from ID "yournamehere"...
aws_s3_bucket_versioning.example: Import prepared!
  Prepared aws_s3_bucket_versioning for import
aws_s3_bucket_versioning.example: Refreshing state... [id=yournamehere]

Import successful!

The resources that were imported are shown above. These resources are now in
your Terraform state and will henceforth be managed by Terraform.
```

#### Ensure Objects Depend on Versioning

When you create an object whose `version_id` you need and an `aws_s3_bucket_versioning` resource in the same configuration, you are more likely to have success by ensuring the `s3_object` depends either implicitly (see below) or explicitly (i.e., using `depends_on = [aws_s3_bucket_versioning.example]`) on the `aws_s3_bucket_versioning` resource.

~> **NOTE:** For critical and/or production S3 objects, do not create a bucket, enable versioning, and create an object in the bucket within the same configuration. Doing so will not allow the AWS-recommended 15 minutes between enabling versioning and writing to the bucket.

This example shows the `aws_s3_object.example` depending implicitly on the versioning resource through the reference to `aws_s3_bucket_versioning.example.bucket` to define `bucket`:

```terraform
resource "aws_s3_bucket" "example" {
  bucket = "yotto"
}

resource "aws_s3_bucket_versioning" "example" {
  bucket = aws_s3_bucket.example.id

  versioning_configuration {
    status = "Enabled"
  }
}

resource "aws_s3_object" "example" {
  bucket = aws_s3_bucket_versioning.example.bucket
  key    = "droeloe"
  source = "example.txt"
}
```

### `website`, `website_domain`, and `website_endpoint` Arguments

Switch your Terraform configuration to the [`aws_s3_bucket_website_configuration` resource](/docs/providers/aws/r/s3_bucket_website_configuration.html) instead.

For example, given this previous configuration:

```terraform
resource "aws_s3_bucket" "example" {
  bucket = "yournamehere"

  # ... other configuration ...
  website {
    index_document = "index.html"
    error_document = "error.html"
  }
}
```

You will get the following error after upgrading:

```
│ Error: Value for unconfigurable attribute
│
│   with aws_s3_bucket.example,
│   on main.tf line 1, in resource "aws_s3_bucket" "example":
│    1: resource "aws_s3_bucket" "example" {
│
│ Can't configure a value for "website": its value will be decided automatically based on the result of applying this configuration.
```

Since `website` is now read only, update your configuration to use the `aws_s3_bucket_website_configuration`
resource and remove `website` and its nested arguments in the `aws_s3_bucket` resource:

```terraform
resource "aws_s3_bucket" "example" {
  bucket = "yournamehere"

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

Run `terraform import` on each new resource, _e.g._,

```shell
$ terraform import aws_s3_bucket_website_configuration.example yournamehere
aws_s3_bucket_website_configuration.example: Importing from ID "yournamehere"...
aws_s3_bucket_website_configuration.example: Import prepared!
  Prepared aws_s3_bucket_website_configuration for import
aws_s3_bucket_website_configuration.example: Refreshing state... [id=yournamehere]

Import successful!

The resources that were imported are shown above. These resources are now in
your Terraform state and will henceforth be managed by Terraform.
```

For example, if you use the `aws_s3_bucket` attribute `website_domain` with `aws_route53_record`, as shown below, you will need to update your configuration:

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

Instead, you will now use the `aws_s3_bucket_website_configuration` resource and its `website_domain` attribute:

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

To delete the default subnet, the above configuration should be updated as follows:

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

Previously, you could set `ecs_target.0.launch_type` to `""`. However, the value `""` is no longer valid. Now, set the argument to `null` (_e.g._, `launch_type = null`) or remove the empty-string configuration.

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

We fix this configuration by setting `launch_type` to `null`:

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

Previously, you could set `ip_address` to `""`, which would result in an AWS error. However, the provider now also gives an error.

### Resource: aws_default_network_acl

Previously, you could set `egress.*.cidr_block`, `egress.*.ipv6_cidr_block`, `ingress.*.cidr_block`, or `ingress.*.ipv6_cidr_block` to `""`. However, the value `""` is no longer valid. Now, set the argument to `null` (_e.g._, `ipv6_cidr_block = null`) or remove the empty-string configuration.

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

To fix this configuration, we remove the empty-string configuration:

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

Previously, you could set `route.*.cidr_block` or `route.*.ipv6_cidr_block` to `""`. However, the value `""` is no longer valid. Now, set the argument to `null` (_e.g._, `ipv6_cidr_block = null`) or remove the empty-string configuration.

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

We fix this configuration by using `null` instead of an empty string (`""`):

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

Previously, you could set `ipv6_cidr_block` to `""`. However, the value `""` is no longer valid. Now, set the argument to `null` (_e.g._, `ipv6_cidr_block = null`) or remove the empty-string configuration.

### Resource: aws_instance

Previously, you could set `private_ip` to `""`. However, the value `""` is no longer valid. Now, set the argument to `null` (_e.g._, `private_ip = null`) or remove the empty-string configuration.

For example, this type of configuration is now not valid:

```terraform
resource "aws_instance" "example" {
  instance_type = "t2.micro"
  private_ip    = ""
}
```

We fix this configuration by removing the empty-string configuration:

```terraform
resource "aws_instance" "example" {
  instance_type = "t2.micro"
}
```

### Resource: aws_efs_mount_target

Previously, you could set `ip_address` to `""`. However, the value `""` is no longer valid. Now, set the argument to `null` (_e.g._, `ip_address = null`) or remove the empty-string configuration.

For example, this type of configuration is now not valid: `ip_address = ""`.

### Resource: aws_elasticsearch_domain

Previously, you could set `ebs_options.0.volume_type` to `""`. However, the value `""` is no longer valid. Now, set the argument to `null` (_e.g._, `volume_type = null`) or remove the empty-string configuration.

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

We fix this configuration by using `null` instead of `""`:

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

We fix this configuration by removing the empty-string configuration:

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

We fix this configuration by using `null` instead of an empty-string (`""`):

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

We fix this configuration by usingd `null` instead of an empty-string (`""`):

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

We fix this configuration by removing `ipv6_cidr_block`:

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

The `aws_subnet_ids` data source has been deprecated and will be removed in a future version. Use the `aws_subnets` data source instead.

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

Version 4.x deprecates the `aws_s3_bucket_object` data source. Maintainers will remove it in a future version. Use `aws_s3_object` instead, where new features and fixes will be added.

## Data Source: aws_s3_bucket_objects

Version 4.x deprecates the `aws_s3_bucket_objects` data source. Maintainers will remove it in a future version. Use `aws_s3_objects` instead, where new features and fixes will be added.

## Resource: aws_batch_compute_environment

You can no longer specify `compute_resources` when `type` is `UNMANAGED`.

Previously, you could apply this configuration and the provider would ignore any compute resources:

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

Now, this configuration is invalid and will result in an error during plan.

To resolve this error, simply remove or comment out the `compute_resources` configuration block.

```hcl
resource "aws_batch_compute_environment" "test" {
  compute_environment_name = "test"

  service_role = aws_iam_role.batch_service.arn
  type         = "UNMANAGED"
}
```

## Resource: aws_cloudwatch_event_target

### Removal of `ecs_target` `launch_type` default value

Previously, the provider assigned `ecs_target` `launch_type` the default value of `EC2` if you did not configure a value. However, the provider no longer assigns a default value.

For example, previously you could workaround the default value by using an empty string (`""`), as shown:

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

This is no longer necessary. We fix the configuration by removing the empty string assignment:

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

Previously, when you did not specify either `engine` or `replication_group_id`, Terraform would not prevent you from applying the invalid configuration.
Now, this will produce an error similar to the one below:

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

Update your configuration to supply one of `engine` or `replication_group_id`.

## Resource: aws_elasticache_global_replication_group

### actual_engine_version Attribute removal

Switch your Terraform configuration from using `actual_engine_version` to use the `engine_version_actual` attribute instead.

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

## Resource: aws_fsx_ontap_storage_virtual_machine

We removed the misspelled argument `active_directory_configuration.0.self_managed_active_directory_configuration.0.organizational_unit_distinguidshed_name` that we previously deprecated. Use `active_directory_configuration.0.self_managed_active_directory_configuration.0.organizational_unit_distinguished_name` now instead. Terraform will automatically migrate the state to `active_directory_configuration.0.self_managed_active_directory_configuration.0.organizational_unit_distinguished_name` during planning.

## Resource: aws_lb_target_group


For `protocol = "TCP"`, you can no longer set `stickiness.type` to `lb_cookie` even when `enabled = false`. Instead, either change the `protocol` to `"HTTP"` or `"HTTPS"`, or change `stickiness.type` to `"source_ip"`.

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

## Resource: aws_s3_bucket_object

Version 4.x deprecates the `aws_s3_bucket_object` and maintainers will remove it in a future version. Use `aws_s3_object` instead, where new features and fixes will be added.

When replacing `aws_s3_bucket_object` with `aws_s3_object` in your configuration, on the next apply, Terraform will recreate the object. If you prefer to not have Terraform recreate the object, import the object using `aws_s3_object`.

For example, the following will import an S3 object into state, assuming the configuration exists, as `aws_s3_object.example`:

```console
% terraform import aws_s3_object.example s3://some-bucket-name/some/key.txt
```

~> **CAUTION:** We do not recommend modifying the state file manually. If you do, you can make it unusable. However, if you accept that risk, some community members have upgraded to the new resource by searching and replacing `"type": "aws_s3_bucket_object",` with `"type": "aws_s3_object",` in the state file, and then running `terraform apply -refresh-only`.

## Resource: aws_spot_instance_request

### instance_interruption_behaviour Argument removal

Switch your Terraform configuration from the `instance_interruption_behaviour` attribute to the `instance_interruption_behavior` attribute instead.

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
