---
subcategory: "AMP (Managed Prometheus)"
layout: "aws"
page_title: "AWS: aws_prometheus_scraper_logging_configuration"
description: |-
  Lists AMP (Managed Prometheus) Scraper Logging Configuration resources.
---

# List Resource: aws_prometheus_scraper_logging_configuration

Lists AMP (Managed Prometheus) Scraper Logging Configuration resources.

## Example Usage

```terraform
list "aws_prometheus_scraper_logging_configuration" "example" {
  provider = aws
}
```

## Argument Reference

This list resource supports the following arguments:

* `region` - (Optional) Region to query. Defaults to provider region.
