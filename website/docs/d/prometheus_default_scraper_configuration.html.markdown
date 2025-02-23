---
subcategory: "AMP (Managed Prometheus)"
layout: "aws"
page_title: "AWS: aws_prometheus_default_scraper_configuration"
description: |-
  Returns the default scraper configuration used when Amazon EKS creates a scraper for you.
---


# Data Source: aws_prometheus_default_scraper_configuration

Returns the default scraper configuration used when Amazon EKS creates a scraper for you.

## Example Usage

```terraform
data "aws_prometheus_default_scraper_configuration" "example" {}
```

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `configuration` - The configuration file.
