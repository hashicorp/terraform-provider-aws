---
subcategory: "CloudWatch"
layout: "aws"
page_title: "AWS: aws_cloudwatch_otel_enrichment"
description: |-
  Manages AWS CloudWatch OTel enrichment.
---

# Resource: aws_cloudwatch_otel_enrichment

Manages AWS CloudWatch OTel enrichment. This is a singleton resource that enables OTel enrichment at the account level.

~> **NOTE:** This resource requires the `aws_observabilityadmin_telemetry_enrichment` resource to be configured first. Without telemetry enrichment enabled, OTel enrichment will not function properly even if the API accepts the configuration.

## Example Usage

### Enable OTel Enrichment

```terraform
resource "aws_observabilityadmin_telemetry_enrichment" "example" {
}

resource "aws_cloudwatch_otel_enrichment" "example" {
  depends_on = [aws_observabilityadmin_telemetry_enrichment.example]
}
```

### Filter Enrichment by Namespace and Metric

By default every namespace that CloudWatch supports is enriched. Narrow that scope with `include_filters` and `exclude_filters`. The configuration below enriches all of `AWS/EC2` except two network metrics, plus exactly two `AWS/RDS` metrics, and nothing else.

```terraform
resource "aws_observabilityadmin_telemetry_enrichment" "example" {
}

resource "aws_cloudwatch_otel_enrichment" "example" {
  include_filters {
    namespace = "AWS/EC2"
  }

  include_filters {
    namespace    = "AWS/RDS"
    metric_names = ["CPUUtilization", "DatabaseConnections"]
  }

  exclude_filters {
    namespace    = "AWS/EC2"
    metric_names = ["NetworkPacketsIn", "NetworkPacketsOut"]
  }

  depends_on = [aws_observabilityadmin_telemetry_enrichment.example]
}
```

## Argument Reference

The following arguments are optional:

* `exclude_filters` - (Optional) Namespaces and metric names to leave unenriched. [See below](#exclude_filters-block).
* `include_filters` - (Optional) Namespaces and metric names to enrich. [See below](#include_filters-block).
* `region` - (Optional) AWS region where this resource is managed.

Filters are evaluated as follows:

* `include_filters` is applied first, then the `exclude_filters` match set is subtracted from the result. Exclusion always wins, so a metric matched by both is not enriched.
* An inactive direction is permissive. With no `include_filters`, every supported namespace is in scope; with no `exclude_filters`, nothing is removed.
* Namespaces and metric names are matched exactly and case-sensitively. There are no wildcards or prefixes, so `AWS/EC2` matches only `AWS/EC2`.
* At most 100 selectors are allowed across `include_filters` and `exclude_filters` **combined**. Two lists of 60 selectors each satisfy the per-list limit but are rejected together.

### `exclude_filters` Block

* `metric_names` - (Optional) Names of the metrics to select within the namespace, up to 100. Omit to match every metric in the namespace.
* `namespace` - (Required) Namespace of the metrics to select. Must be 1-255 characters and must not begin with a colon.

### `include_filters` Block

* `metric_names` - (Optional) Names of the metrics to select within the namespace, up to 100. Omit to match every metric in the namespace.
* `namespace` - (Required) Namespace of the metrics to select. Must be 1-255 characters and must not begin with a colon.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - (Deprecated) AWS region where the enrichment is managed.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `5m`)
* `update` - (Default `5m`)
* `delete` - (Default `5m`)

## Import

In Terraform v1.12.0 and later, the [`import` block](https://developer.hashicorp.com/terraform/language/import) can be used with the `identity` attribute. For example:

```terraform
import {
  to = aws_cloudwatch_otel_enrichment.example
  identity = {
  }
}

resource "aws_cloudwatch_otel_enrichment" "example" {
}
```

### Identity Schema

#### Required

No required attributes for singleton identity.

#### Optional

* `account_id` (String) AWS Account where this resource is managed.
* `region` (String) Region where this resource is managed.

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import CloudWatch OTel Enrichment using the region. For example:

```terraform
import {
  to = aws_cloudwatch_otel_enrichment.example
  id = "us-west-2"
}
```

Using `terraform import`, import CloudWatch OTel Enrichment using the region. For example:

```console
% terraform import aws_cloudwatch_otel_enrichment.example us-west-2
```
