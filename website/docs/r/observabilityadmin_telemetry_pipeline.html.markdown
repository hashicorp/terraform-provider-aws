---
subcategory: "CloudWatch Observability Admin"
layout: "aws"
page_title: "AWS: aws_observabilityadmin_telemetry_pipeline"
description: |-
  Manages an AWS CloudWatch Observability Admin Telemetry Pipeline.
---

# Resource: aws_observabilityadmin_telemetry_pipeline

Manages an AWS CloudWatch Observability Admin Telemetry Pipeline.

Telemetry pipelines allow you to collect, transform, and route telemetry data from AWS services. Each pipeline defines a source, optional processors, and one or more sinks for the telemetry data.

For more information, see the [AWS CloudWatch Observability Admin Telemetry Pipelines documentation](https://docs.aws.amazon.com/cloudwatch/latest/observabilityadmin/what-is-observabilityadmin.html).

~> **NOTE:** Only one telemetry pipeline per data source type is allowed per account. For example, you can have one pipeline for `amazon_api_gateway/access` and another for `amazon_vpc/flow`, but not two pipelines for the same data source type.

## Example Usage

### Basic Pipeline

```terraform
data "aws_caller_identity" "current" {}
data "aws_partition" "current" {}

resource "aws_iam_role" "example" {
  name = "example-telemetry-pipeline"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Action = "sts:AssumeRole"
      Effect = "Allow"
      Principal = {
        Service = "observabilityadmin.amazonaws.com"
      }
    }]
  })
}

resource "aws_iam_role_policy" "example" {
  role = aws_iam_role.example.name

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Effect = "Allow"
      Action = [
        "logs:CreateLogGroup",
        "logs:CreateLogStream",
        "logs:PutLogEvents",
        "logs:DescribeLogGroups",
        "logs:DescribeLogStreams",
      ]
      Resource = "arn:${data.aws_partition.current.partition}:logs:*:${data.aws_caller_identity.current.account_id}:*"
    }]
  })
}

resource "aws_observabilityadmin_telemetry_pipeline" "example" {
  name = "example-pipeline"

  configuration {
    body = yamlencode({
      pipeline = {
        source = {
          cloudwatch_logs = {
            aws = {
              sts_role_arn = aws_iam_role.example.arn
            }
            log_event_metadata = {
              data_source_name = "amazon_api_gateway"
              data_source_type = "access"
            }
          }
        }
        sink = [{
          cloudwatch_logs = {
            log_group = "@original"
          }
        }]
      }
    })
  }

  depends_on = [aws_iam_role_policy.example]
}
```

### Pipeline with Processor

```terraform
resource "aws_observabilityadmin_telemetry_pipeline" "example" {
  name = "example-vpc-pipeline"

  configuration {
    body = yamlencode({
      pipeline = {
        source = {
          cloudwatch_logs = {
            aws = {
              sts_role_arn = aws_iam_role.example.arn
            }
            log_event_metadata = {
              data_source_name = "amazon_vpc"
              data_source_type = "flow"
            }
          }
        }
        processor = [{
          ocsf = {
            schema = {
              vpc_flow = null
            }
            version         = "1.5"
            mapping_version = "1.5.0"
          }
        }]
        sink = [{
          cloudwatch_logs = {
            log_group = "@original"
          }
        }]
      }
    })
  }

  depends_on = [aws_iam_role_policy.example]
}
```

## Argument Reference

This resource supports the following arguments:

* `name` - (Required, Forces new resource) Name of the telemetry pipeline. Must be between 3 and 28 characters, start with a lowercase letter, and contain only lowercase letters, digits, and hyphens.
* `configuration` - (Required) Configuration block for the telemetry pipeline. See [`configuration`](#configuration) below.

The following arguments are optional:

* `tags` - (Optional) Key-value map of resource tags. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

### configuration

* `body` - (Required) The pipeline configuration body. This is a YAML-encoded string defining the pipeline source, optional processors, and sinks.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the telemetry pipeline.
* `created_timestamp` - Unix epoch timestamp (in seconds) when the pipeline was created.
* `last_update_timestamp` - Unix epoch timestamp (in seconds) when the pipeline was last updated.
* `status` - Current status of the pipeline (e.g., `CREATING`, `ACTIVE`, `UPDATING`, `DELETING`).
* `status_reason` - Description of the reason for the current status, if available.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

- `create` - (Default `30m`)
- `update` - (Default `30m`)
- `delete` - (Default `30m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import CloudWatch Observability Admin Telemetry Pipeline using the `name`. For example:

```terraform
import {
  to = aws_observabilityadmin_telemetry_pipeline.example
  id = "example-pipeline"
}
```

Using `terraform import`, import CloudWatch Observability Admin Telemetry Pipeline using the `name`. For example:

```console
% terraform import aws_observabilityadmin_telemetry_pipeline.example example-pipeline
```
