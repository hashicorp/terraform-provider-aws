---
subcategory: "AMP (Managed Prometheus)"
layout: "aws"
page_title: "AWS: aws_prometheus_anomaly_detector"
description: |-
  Manages an AWS AMP (Managed Prometheus) Anomaly Detector.
---

# Resource: aws_prometheus_anomaly_detector

Manages an AWS AMP (Managed Prometheus) Anomaly Detector.

## Example Usage

### Basic Usage

```terraform
resource "aws_prometheus_workspace" "example" {}

resource "aws_prometheus_anomaly_detector" "example" {
  alias        = "example"
  workspace_id = aws_prometheus_workspace.example.id

  configuration {
    random_cut_forest {
      query = "avg(up)"
    }
  }

  missing_data_action {
    skip = true
  }
}
```

### With evaluation interval and labels

```terraform
resource "aws_prometheus_workspace" "example" {}

resource "aws_prometheus_anomaly_detector" "example" {
  alias                          = "example"
  workspace_id                   = aws_prometheus_workspace.example.id
  evaluation_interval_in_seconds = 120

  labels = {
    env  = "production"
    team = "platform"
  }

  configuration {
    random_cut_forest {
      query        = "avg(up)"
      sample_size  = 256
      shingle_size = 4

      ignore_near_expected_from_above {
        ratio = 1.5
      }

      ignore_near_expected_from_below {
        amount = 2.0
      }
    }
  }

  missing_data_action {
    mark_as_anomaly = true
  }
}
```

## Argument Reference

The following arguments are required:

* `alias` - (Required) Name of the anomaly detector.
* `configuration` - (Required) Configuration block for the anomaly detector algorithm. See [`configuration`](#configuration) below.
* `missing_data_action` - (Required) Configuration block for the action to take when data is missing. See [`missing_data_action`](#missing_data_action) below.
* `workspace_id` - (Required) ID of the AMP workspace in which to create the anomaly detector.

The following arguments are optional:

* `evaluation_interval_in_seconds` - (Optional, Computed) Interval in seconds at which the anomaly detector evaluates data.
* `labels` - (Optional) Map of label key-value pairs used to scope the anomaly detector to specific time series.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `tags` - (Optional) Map of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

### `configuration` Block

* `random_cut_forest` - (Required) Configuration block for the Random Cut Forest anomaly detection algorithm. See [`random_cut_forest`](#random_cut_forest) below.

#### `random_cut_forest` Block

* `ignore_near_expected_from_above` - (Optional) Configuration block for suppressing anomalies when the observed value is slightly above the expected value. See [`ignore_near_expected_from_above`](#ignore_near_expected_from_above-and-ignore_near_expected_from_below) below.
* `ignore_near_expected_from_below` - (Optional) Configuration block for suppressing anomalies when the observed value is slightly below the expected value. See [`ignore_near_expected_from_below`](#ignore_near_expected_from_above-and-ignore_near_expected_from_below) below.
* `query` - (Required) PromQL query used to select the time series for anomaly detection.
* `sample_size` - (Optional, Computed) Number of data points used to train the model. Must be at least `256`.
* `shingle_size` - (Optional, Computed) Number of consecutive data points that form a single input to the model. Must be at least `2`.

#### `ignore_near_expected_from_above` Block and `ignore_near_expected_from_below` Block

Exactly one of `amount` or `ratio` must be specified.

* `amount` - (Optional) Absolute amount by which the observed value may exceed the expected value before being reported as an anomaly. Conflicts with `ratio`.
* `ratio` - (Optional) Ratio by which the observed value may exceed the expected value before being reported as an anomaly. Must be at least `0`. Conflicts with `amount`.

### `missing_data_action` Block

Exactly one of `mark_as_anomaly` or `skip` must be specified.

* `mark_as_anomaly` - (Optional) Whether to treat missing data points as anomalies. Must be set to `true`. Conflicts with `skip`.
* `skip` - (Optional) Whether to skip missing data points without reporting them as anomalies. Must be set to `true`. Conflicts with `mark_as_anomaly`.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the Anomaly Detector.
* `created_at` - RFC3339 timestamp of when the anomaly detector was created.
* `id` - Unique identifier of the anomaly detector.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `10m`)
* `update` - (Default `10m`)
* `delete` - (Default `10m`)

## Import

In Terraform v1.12.0 and later, the [`import` block](https://developer.hashicorp.com/terraform/language/import) can be used with the `identity` attribute. For example:

```terraform
import {
  to = aws_prometheus_anomaly_detector.example
  identity = {
    id           = "ad-12345678-abcd-1234-abcd-123456789012"
    workspace_id = "ws-12345678-abcd-1234-abcd-123456789012"
  }
}

resource "aws_prometheus_anomaly_detector" "example" {
  ### Configuration omitted for brevity ###
}
```

### Identity Schema

#### Required

* `id` - (String) ID of the Anomaly Detector.
* `workspace_id` - (String) ID of the AMP workspace containing the Anomaly Detector.

#### Optional

* `account_id` (String) AWS Account where this resource is managed.
* `region` (String) Region where this resource is managed.

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import AMP (Managed Prometheus) Anomaly Detector using a comma-delimited string combining `id` and `workspace_id`. For example:

```terraform
import {
  to = aws_prometheus_anomaly_detector.example
  id = "ad-12345678-abcd-1234-abcd-123456789012,ws-12345678-abcd-1234-abcd-123456789012"
}
```

Using `terraform import`, import AMP (Managed Prometheus) Anomaly Detector using a comma-delimited string combining `id` and `workspace_id`. For example:

```console
% terraform import aws_prometheus_anomaly_detector.example ad-12345678-abcd-1234-abcd-123456789012,ws-12345678-abcd-1234-abcd-123456789012
```
