---
subcategory: "ECR (Elastic Container Registry)"
layout: "aws"
page_title: "AWS: aws_ecr_registry_scanning_configuration"
description: |-
  Provides an Elastic Container Registry Scanning Configuration.
---

# Resource: aws_ecr_registry_scanning_configuration

Provides an Elastic Container Registry Scanning Configuration. Can't be completely deleted, instead reverts to the default `BASIC` scanning configuration without rules.

## Example Usage

### Basic example

```terraform
resource "aws_ecr_registry_scanning_configuration" "configuration" {
  scan_type = "ENHANCED"

  rule {
    scan_frequency = "CONTINUOUS_SCAN"
    repository_filter {
      filter      = "example"
      filter_type = "WILDCARD"
    }
  }
}
```

### Multiple rules

```terraform
resource "aws_ecr_registry_scanning_configuration" "test" {
  scan_type = "ENHANCED"

  rule {
    scan_frequency = "SCAN_ON_PUSH"
    repository_filter {
      filter      = "*"
      filter_type = "WILDCARD"
    }
  }

  rule {
    scan_frequency = "CONTINUOUS_SCAN"
    repository_filter {
      filter      = "example"
      filter_type = "WILDCARD"
    }
  }
}
```

## Argument Reference

This resource supports the following arguments:

- `scan_type` - (Required) the scanning type to set for the registry. Can be either `ENHANCED` or `BASIC`.
- `rule` - (Optional) One or multiple blocks specifying scanning rules to determine which repository filters are used and at what frequency scanning will occur. See [below for schema](#rule).

### rule

- `repository_filter` - (Required) One or more repository filter blocks, containing a `filter` (required string filtering repositories, see pattern regex [here](https://docs.aws.amazon.com/AmazonECR/latest/APIReference/API_ScanningRepositoryFilter.html)) and a `filter_type` (required string, currently only `WILDCARD` is supported).
- `scan_frequency` - (Required) The frequency that scans are performed at for a private registry. Can be `SCAN_ON_PUSH`, `CONTINUOUS_SCAN`, or `MANUAL`.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `registry_id` - The registry ID the scanning configuration applies to.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import ECR Scanning Configurations using the `registry_id`. For example:

```terraform
import {
  to = aws_ecr_registry_scanning_configuration.example
  id = "123456789012"
}
```

Using `terraform import`, import ECR Scanning Configurations using the `registry_id`. For example:

```console
% terraform import aws_ecr_registry_scanning_configuration.example 123456789012
```
