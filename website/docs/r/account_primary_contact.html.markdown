---
subcategory: "Account Management"
layout: "aws"
page_title: "AWS: aws_account_primary_contact"
description: |-
  Manages the specified primary contact information associated with an AWS Account.
---

# Resource: aws_account_primary_contact

Manages the specified primary contact information associated with an AWS Account.

## Example Usage

```terraform
resource "aws_account_primary_contact" "test" {
  address_line_1     = "123 Any Street"
  city               = "Seattle"
  company_name       = "Example Corp, Inc."
  country_code       = "US"
  district_or_county = "King"
  full_name          = "My Name"
  phone_number       = "+64211111111"
  postal_code        = "98101"
  state_or_region    = "WA"
  website_url        = "https://www.examplecorp.com"
}
```

## Argument Reference

This resource supports the following arguments:

* `account_id` - (Optional) The ID of the target account when managing member accounts. Will manage current user's account by default if omitted.
* `address_line_1` - (Required) The first line of the primary contact address.
* `address_line_2` - (Optional) The second line of the primary contact address, if any.
* `address_line_3` - (Optional) The third line of the primary contact address, if any.
* `city` - (Required) The city of the primary contact address.
* `company_name` - (Optional) The name of the company associated with the primary contact information, if any.
* `country_code` - (Required) The ISO-3166 two-letter country code for the primary contact address.
* `district_or_county` - (Optional) The district or county of the primary contact address, if any.
* `full_name` - (Required) The full name of the primary contact address.
* `phone_number` - (Required) The phone number of the primary contact information. The number will be validated and, in some countries, checked for activation.
* `postal_code` - (Required) The postal code of the primary contact address.
* `state_or_region` - (Optional) The state or region of the primary contact address. This field is required in selected countries.
* `website_url` - (Optional) The URL of the website associated with the primary contact information, if any.

## Attribute Reference

This resource exports no additional attributes.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import the Primary Contact using the `account_id`. For example:

```terraform
import {
  to = aws_account_primary_contact.test
  id = "1234567890"
}
```

Using `terraform import`, import the Primary Contact using the `account_id`. For example:

```console
% terraform import aws_account_primary_contact.test 1234567890
```
