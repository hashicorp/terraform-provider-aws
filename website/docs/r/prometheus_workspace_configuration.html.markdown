---
subcategory: "AMP (Managed Prometheus)"
layout: "aws"
page_title: "AWS: aws_prometheus_workspace_configuration"
description: |-
  Terraform resource for managing an AWS Managed Service for Prometheus Workspace Configuration.
---
# Resource: aws_prometheus_workspace_configuration

Manages an AWS Managed Service for Prometheus Workspace Configuration.

## Example Usage

### Basic Usage

```terraform
resource "aws_prometheus_workspace" "example" {}

resource "aws_prometheus_workspace_configuration" "example" {
  workspace_id             = aws_prometheus_workspace.example.id
  retention_period_in_days = 60

  limits_per_label_set {
    label_set = {
      "env" = "dev"
    }
    limits {
      max_series = 100000
    }
  }

  limits_per_label_set {
    label_set = {
      "env" = "prod"
    }
    limits {
      max_series = 400000
    }
  }
}
```

### Setting up default bucket

The default bucket limit is the maximum number of active time series that can be
ingested in the workspace, counting only time series that don’t match a defined
label set.

```terraform
resource "aws_prometheus_workspace" "example" {}

resource "aws_prometheus_workspace_configuration" "example" {
  workspace_id = aws_prometheus_workspace.example.id

  limits_per_label_set {
    label_set = {}
    limits {
      max_series = 50000
    }
  }
}
```

### With out-of-order and rule query configuration

```terraform
resource "aws_prometheus_workspace" "example" {}

resource "aws_prometheus_workspace_configuration" "example" {
  workspace_id                          = aws_prometheus_workspace.example.id
  retention_period_in_days              = 30
  out_of_order_time_window_in_seconds   = 120
  rule_query_offset_in_seconds          = 300
}
```

## Argument Reference

The following arguments are required:

* `workspace_id` - (Required) ID of the workspace to configure.

The following arguments are optional:

* `limits_per_label_set` - (Optional) Configuration block for setting limits on metrics with specific label sets. Detailed below.
* `out_of_order_time_window_in_seconds` - (Optional) Time window in seconds for accepting out-of-order samples. Must be between 0 and 600 seconds.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `retention_period_in_days` - (Optional) Number of days to retain metric data in the workspace.
* `rule_query_offset_in_seconds` - (Optional) Query offset in seconds for rule evaluation. Must be between 0 and 86400 seconds.

### `limits_per_label_set`

The `limits_per_label_set` configuration block supports the following arguments:

* `label_set` - (Required) Map of label key-value pairs that identify the metrics to which the limits apply. An empty map represents the default bucket for metrics that don't match any other label set.
* `limits` - (Required) Configuration block for the limits to apply to the specified label set. Detailed below.

#### `limits`

The `limits` configuration block supports the following arguments:

* `max_series` - (Required) Maximum number of active time series that can be ingested for metrics matching the label set.

## Attribute Reference

This resource exports no additional attributes.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `5m`)
* `update` - (Default `5m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import AMP (Managed Prometheus) Workspace Configuration using the `workspace_id`. For example:

```terraform
import {
  to = aws_prometheus_workspace_configuration.example
  id = "ws-12345678-abcd-1234-abcd-123456789012"
}
```

Using `terraform import`, import AMP (Managed Prometheus) Workspace Configuration using the `workspace_id`. For example

```console
% terraform import aws_prometheus_workspace_configuration.example ws-12345678-abcd-1234-abcd-123456789012
```
