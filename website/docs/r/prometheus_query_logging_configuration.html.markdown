---
subcategory: "AMP (Managed Prometheus)"
layout: "aws"
page_title: "AWS: aws_prometheus_query_logging_configuration"
description: |-
  Manages an Amazon Managed Service for Prometheus (AMP) Query Logging Configuration.
---

# Resource: aws_prometheus_query_logging_configuration

Manages an Amazon Managed Service for Prometheus (AMP) Query Logging Configuration.

## Example Usage

```terraform
resource "aws_prometheus_workspace" "example" {
  alias = "example"
}

resource "aws_cloudwatch_log_group" "example" {
  name = "/aws/prometheus/query-logs/example"
}

resource "aws_prometheus_query_logging_configuration" "example" {
  workspace_id = aws_prometheus_workspace.example.id

  destination {
    cloudwatch_logs {
      log_group_arn = "${aws_cloudwatch_log_group.example.arn}:*"
    }

    filters {
      qsp_threshold = 1000
    }
  }
}
```

## Argument Reference

This resource supports the following arguments:

* `destination` - (Required) Configuration block for the logging destinations. See [`destinations`](#destinations).
* `workspace_id` - (Required) The ID of the AMP workspace for which to configure query logging.

The following arguments are optional:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

### `destination`

* `cloudwatch_logs` - (Required) Configuration block for CloudWatch Logs destination. See [`cloudwatch_logs`](#cloudwatch_logs).
* `filters` - (Required) A list of filter configurations that specify which logs should be sent to the destination. See [`filters`](#filters).

#### `cloudwatch_logs`

* `log_group_arn` - (Required) The ARN of the CloudWatch log group to which query logs will be sent. The ARN must end with `:*`

#### `filters`

* `qsp_threshold` - (Required) The Query Samples Processed (QSP) threshold above which queries will be logged. Queries processing more samples than this threshold will be captured in logs.

## Attribute Reference

This resource exports no additional attributes.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import the Query Logging Configuration using the workspace ID. For example:

```terraform
import {
  to = aws_prometheus_query_logging_configuration.example
  id = "ws-12345678-90ab-cdef-1234-567890abcdef"
}
```

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

- `create` - (Default `5m`)
- `update` - (Default `5m`)
- `delete` - (Default `5m`)
