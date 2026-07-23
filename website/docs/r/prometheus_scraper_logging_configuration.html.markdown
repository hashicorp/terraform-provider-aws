---
subcategory: "AMP (Managed Prometheus)"
layout: "aws"
page_title: "AWS: aws_prometheus_scraper_logging_configuration"
description: |-
  Manages an Amazon Managed Service for Prometheus (AMP) Scraper Logging Configuration.
---

# Resource: aws_prometheus_scraper_logging_configuration

Manages an Amazon Managed Service for Prometheus (AMP) Scraper Logging Configuration.

## Example Usage

### Basic Usage

```terraform
resource "aws_prometheus_scraper" "example" {
  source {
    eks {
      cluster_arn = aws_eks_cluster.example.arn
      subnet_ids  = aws_subnet.example[*].id
    }
  }

  destination {
    amp {
      workspace_arn = aws_prometheus_workspace.example.arn
    }
  }

  scrape_configuration = <<EOT
global:
  scrape_interval: 15s
scrape_configs:
  - job_name: 'prometheus'
    static_configs:
      - targets: ['localhost:9090']
EOT
}

resource "aws_cloudwatch_log_group" "example" {
  name = "/aws/prometheus/scraper-logs/example"
}

resource "aws_prometheus_scraper_logging_configuration" "example" {
  scraper_id = aws_prometheus_scraper.example.id

  logging_destination {
    cloudwatch_logs {
      log_group_arn = "${aws_cloudwatch_log_group.example.arn}:*"
    }
  }
}
```

### With Scraper Components

```terraform
resource "aws_prometheus_scraper_logging_configuration" "example" {
  scraper_id = aws_prometheus_scraper.example.id

  scraper_components {
    type = "COLLECTOR"
  }

  scraper_components {
    type = "EXPORTER"
    options = {
      "log_level" = "debug"
    }
  }

  logging_destination {
    cloudwatch_logs {
      log_group_arn = "${aws_cloudwatch_log_group.example.arn}:*"
    }
  }
}
```

## Argument Reference

This resource supports the following arguments:

* `logging_destination` - (Required) Configuration block for the logging destination. See [`logging_destination`](#logging_destination).
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `scraper_id` - (Required) The ID of the scraper to configure logging for.

The following arguments are optional:

* `scraper_components` - (Optional) Configuration blocks for scraper components to log. See [`scraper_components`](#scraper_components).

### `logging_destination`

* `cloudwatch_logs` - (Required) Configuration block for CloudWatch Logs destination. See [`cloudwatch_logs`](#cloudwatch_logs).

### `cloudwatch_logs`

* `log_group_arn` - (Required) The ARN of the CloudWatch Logs log group. Must end with `:*`.

### `scraper_components`

* `type` - (Required) The type of scraper component. Valid values: `COLLECTOR`, `EXPORTER`.
* `options` - (Optional) Map of configuration options for the scraper component.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `scraper_id` - The ID of the scraper.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `5m`)
* `update` - (Default `5m`)
* `delete` - (Default `5m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import AMP Scraper Logging Configuration using the `scraper_id`. For example:

```terraform
import {
  to = aws_prometheus_scraper_logging_configuration.example
  id = "s-example1234567890abcdef0"
}
```

Using `terraform import`, import AMP Scraper Logging Configuration using the `scraper_id`. For example:

```console
% terraform import aws_prometheus_scraper_logging_configuration.example s-example1234567890abcdef0
```
