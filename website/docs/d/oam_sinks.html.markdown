---
subcategory: "CloudWatch Observability Access Manager"
layout: "aws"
page_title: "AWS: aws_oam_sinks"
description: |-
  Terraform data source for managing an AWS CloudWatch Observability Access Manager Sinks.
---

# Data Source: aws_oam_sinks

Terraform data source for managing an AWS CloudWatch Observability Access Manager Sinks.

## Example Usage

### Basic Usage

```terraform
data "aws_oam_sinks" "example" {
}
```

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `arns` - Set of ARN of the Sinks.
