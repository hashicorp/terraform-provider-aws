---
subcategory: "Tax Settings"
layout: "aws"
page_title: "AWS: aws_taxsettings_tax_registration"
description: |-
  Manages tax registration information for an AWS account.
---

# Resource: aws_taxsettings_tax_registration

Manages tax registration information (TRN) for an AWS account. Tax registration is a singleton per account — creating this resource when one already exists updates it in place.

## Example Usage

### Basic VAT Registration

```terraform
resource "aws_taxsettings_tax_registration" "example" {
  registration_id   = "GB123456789"
  registration_type = "VAT"
  legal_name        = "Example Company Ltd"

  legal_address {
    address_line1 = "123 Any Street"
    city          = "London"
    country_code  = "GB"
    postal_code   = "EC1A 1BB"
  }
}
```

### GST Registration with Optional Fields

```terraform
resource "aws_taxsettings_tax_registration" "example" {
  registration_id   = "12345678901"
  registration_type = "GST"
  legal_name        = "Example Pty Ltd"
  sector            = "Business"

  legal_address {
    address_line1   = "100 George Street"
    city            = "Sydney"
    country_code    = "AU"
    postal_code     = "2000"
    state_or_region = "NSW"
  }
}
```

### Setting Tax Registration for a Member Account

```terraform
resource "aws_taxsettings_tax_registration" "member" {
  account_id        = "123456789012"
  registration_id   = "DE123456789"
  registration_type = "VAT"
  legal_name        = "Tochtergesellschaft GmbH"

  legal_address {
    address_line1 = "Beispielstraße 1"
    city          = "Berlin"
    country_code  = "DE"
    postal_code   = "10115"
  }
}
```

## Argument Reference

The following arguments are required:

* `registration_id` - (Required) Tax registration number (TRN), such as a VAT number, GST number, or ABN. Maximum 200 characters.
* `registration_type` - (Required) Type of tax registration. Valid values: `VAT`, `GST`, `CPF`, `CNPJ`, `SST`, `TIN`, `NRIC`.

The following arguments are optional:

* `account_id` - (Optional, Forces new resource) AWS account ID to set the tax registration for. Defaults to the account of the calling principal. Use this when managing member accounts from a management account.
* `certified_email_id` - (Optional) Email address to receive VAT invoices. Required for South Korea.
* `legal_address` - (Optional) Legal address associated with the tax registration. Required for most countries. See [`legal_address`](#legal_address) below.
* `legal_name` - (Optional) Legal name of the business associated with the registration. Required for most countries; not required for Brazil CNPJ registrations.
* `sector` - (Optional) Business sector. Valid values: `Business`, `Individual`, `Government`. Required for Turkey.

### legal_address

* `address_line1` - (Required) First line of the address.
* `city` - (Required) City.
* `country_code` - (Required) ISO 3166-1 alpha-2 country code.
* `postal_code` - (Required) Postal or ZIP code.
* `address_line2` - (Optional) Second line of the address.
* `address_line3` - (Optional) Third line of the address. Required for Saudi Arabia (building number).
* `district_or_county` - (Optional) District or county. Required for Turkey; used for neighborhood in Brazil.
* `state_or_region` - (Optional) State or region. Required for Canada, India, UAE, Romania, and Brazil CPF registrations.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - AWS account ID that the registration belongs to.
* `status` - Current status of the tax registration. One of `Verified`, `Pending`, `Deleted`, `Rejected`. Newly created registrations are typically `Pending` until AWS verifies them.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import a Tax Registration using the AWS account ID. For example:

```terraform
import {
  to = aws_taxsettings_tax_registration.example
  id = "123456789012"
}
```

Using `terraform import`, import a Tax Registration using the AWS account ID. For example:

```console
% terraform import aws_taxsettings_tax_registration.example 123456789012
```
