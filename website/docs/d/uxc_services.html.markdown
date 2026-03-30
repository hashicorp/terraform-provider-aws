---
subcategory: "User Experience Customization"
layout: "aws"
page_title: "AWS: aws_uxc_services"
description: |-
  Data source for retrieving the list of AWS services available in UXC.
---

# Data Source: aws_uxc_services

Use this data source to retrieve the list of AWS service identifiers available for use with the [`aws_uxc_account_customizations`](../r/uxc_account_customizations.html.markdown) resource's `visible_services` attribute.

~> **Note:** This data source operates globally and always queries the `us-east-1` region regardless of the provider region configuration.

## Example Usage

### List All Available Services

```terraform
data "aws_uxc_services" "example" {}

output "available_services" {
  value = data.aws_uxc_services.example.services
}
```

### Use with Account Customizations

```terraform
data "aws_uxc_services" "example" {}

resource "aws_uxc_account_customizations" "example" {
  visible_services = data.aws_uxc_services.example.services
}
```

## Argument Reference

This data source does not support any arguments.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `services` - List of AWS service identifiers available in UXC.
