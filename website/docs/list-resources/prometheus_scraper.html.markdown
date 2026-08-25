---
subcategory: "AMP (Managed Prometheus)"
layout: "aws"
page_title: "AWS: aws_prometheus_scraper"
description: |-
  Lists AMP (Managed Prometheus) Scraper resources.
---

# List Resource: aws_prometheus_scraper

Lists AMP (Managed Prometheus) Scraper resources.

## Example Usage

```terraform
list "aws_prometheus_scraper" "example" {
  provider = aws
}
```

### With Filters

```terraform
list "aws_prometheus_scraper" "example" {
  provider = aws

  config {
    filters = {
      status = ["ACTIVE", "CREATING"]
    }
  }
}
```

## Argument Reference

This list resource supports the following arguments:

* `filters` - (Optional) List of key-value pairs to filter the list of scrapers returned.
* `region` - (Optional) Region to query. Defaults to provider region.
