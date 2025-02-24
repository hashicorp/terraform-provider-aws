---
subcategory: "Macie"
layout: "aws"
page_title: "AWS: aws_macie2_organization_configuration"
description: |-
  Provides a resource to manage Amazon Macie configuration settings for an organization in AWS Organizations.
---

# Resource: aws_macie2_organization_configuration

Provides a resource to manage Amazon Macie configuration settings for an organization in AWS Organizations.

## Example Usage

```terraform
resource "aws_macie2_organization_configuration" "example" {
  auto_enable = true
}
```

## Argument Reference

This resource supports the following arguments:

* `auto_enable` - (Required) Whether to enable Amazon Macie automatically for accounts that are added to the organization in AWS Organizations.

## Attribute Reference

This resource exports no additional attributes.
