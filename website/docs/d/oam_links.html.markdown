---
subcategory: "CloudWatch Observability Access Manager"
layout: "aws"
page_title: "AWS: aws_oam_links"
description: |-
  Terraform data source for managing an AWS CloudWatch Observability Access Manager Links.
---

# Data Source: aws_oam_links

Terraform data source for managing an AWS CloudWatch Observability Access Manager Links.

## Example Usage

### Basic Usage

```terraform
data "aws_oam_links" "example" {
}
```

## Attributes Reference

The following attributes are exported:

* `arns` - Set of ARN of the Links.
