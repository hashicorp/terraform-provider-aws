---
subcategory: "DevOps Guru"
layout: "aws"
page_title: "AWS: aws_devopsguru_event_sources_config"
description: |-
  Terraform resource for managing an AWS DevOps Guru Event Sources Config.
---
# Resource: aws_devopsguru_event_sources_config

Terraform resource for managing an AWS DevOps Guru Event Sources Config. Currently the only service that can be integrated with DevOps Guru is Amazon CodeGuru Profiler, which can produce proactive recommendations which can be stored and viewed in DevOps Guru.

~> Destruction of this resource will set the CodeGuru profiler status to `DISABLED`. If you wish to preserve an `ENABLED` configuration while removing the Terraform resource, utilize a [`removed` block](https://developer.hashicorp.com/terraform/language/resources/syntax#removing-resources) (available in Terraform 1.7+).

~> Event sources are configured at the account level. To avoid persistent differences, this resource should be defined only once.

## Example Usage

### Basic Usage

```terraform
resource "aws_devopsguru_event_sources_config" "example" {
  event_sources {
    amazon_code_guru_profiler {
      status = "ENABLED"
    }
  }
}
```

## Argument Reference

The following arguments are required:

* `event_sources` - (Required) Configuration information about the integration of DevOps Guru as the Consumer via EventBridge with another AWS Service. See [`event_sources`](#event_sources-argument-reference) below.

### `event_sources` Argument Reference

* `amazon_code_guru_profiler` - (Required) Stores whether DevOps Guru is configured to consume recommendations which are generated from AWS CodeGuru Profiler. See [`amazon_code_guru_profiler`](#amazon_code_guru_profiler-argument-reference) below.

### `amazon_code_guru_profiler` Argument Reference

* `status` - (Required) Status of the CodeGuru Profiler integration. Valid values are `ENABLED` and `DISABLED`.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - AWS region.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import DevOps Guru Event Sources Config using the `id`. For example:

```terraform
import {
  to = aws_devopsguru_event_sources_config.example
  id = "us-east-1"
}
```

Using `terraform import`, import DevOps Guru Event Sources Config using the `id`. For example:

```console
% terraform import aws_devopsguru_event_sources_config.example us-east-1
```
