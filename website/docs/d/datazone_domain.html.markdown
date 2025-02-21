---
subcategory: "DataZone"
layout: "aws"
page_title: "AWS: aws_datazone_domain"
description: |-
  Terraform data source for managing an AWS DataZone Domain.
---

# Data Source: aws_datazone_domain

Terraform data source for managing an AWS DataZone Domain.

## Example Usage

### Basic Usage

```terraform
data "aws_datazone_domain" "example" {
  name = "example_domain"
}
```

## Argument Reference

The following arguments are required:

* `name` - (Required) Name of the Domain.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `arn` - ARN of the Domain.
* `id` - ID of the Domain.
