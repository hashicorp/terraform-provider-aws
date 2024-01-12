---
subcategory: "AMP (Managed Prometheus)"
layout: "aws"
page_title: "AWS: aws_prometheus_scraper_configuration"
description: |-
  Terraform data source for managing an AWS AMP (Managed Prometheus) Scrape Configuration.
---


# Data Source: aws_prometheus_scraper_configuration

Use this data source to get the latest default Prometheus scraper
configuration for AWS managed collectors.

## Example Usage

### Basic Usage

```terraform
data "aws_prometheus_scrape_configuration" "example" {}
```

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `default` - default YAML Prometheus scraper configuration
