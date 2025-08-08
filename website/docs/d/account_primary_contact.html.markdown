---
subcategory: "Account Management"
layout: "aws"
page_title: "AWS: aws_account_primary_contact"
description: |-
  Terraform data source for the primary contact information associated with an AWS Account.
---

# Data Source: aws_account_primary_contact

Terraform data source for the primary contact information associated with an AWS Account.

## Example Usage

```terraform
data "aws_account_primary_contact" "test" {}
```

## Argument Reference

This data source supports the following arguments:

* `account_id` - (Optional) The ID of the target account when managing member accounts. Will manage current user's account by default if omitted.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `address_line_1` - The first line of the primary contact address.
* `address_line_2` - The second line of the primary contact address.
* `address_line_3` - The third line of the primary contact address.
* `city` - The city of the primary contact address.
* `company_name` - The name of the company associated with the primary contact information.
* `country_code` - The ISO-3166 two-letter country code for the primary contact address.
* `district_or_county` - The district or county of the primary contact address.
* `full_name` - The full name of the primary contact address.
* `phone_number` - The phone number of the primary contact information.
* `postal_code` - The postal code of the primary contact address.
* `state_or_region` - The state or region of the primary contact address.
* `website_url` - The URL of the website associated with the primary contact information.
