---
subcategory: "Connect Customer Profiles"
layout: "aws"
page_title: "AWS: aws_customerprofiles_profile"
description: |-
  Terraform resource for managing an Amazon Customer Profiles Profile.
---

# Resource: aws_customerprofiles_profile

Terraform resource for managing an Amazon Customer Profiles Profile.
See the [Create Profile](https://docs.aws.amazon.com/customerprofiles/latest/APIReference/API_CreateProfile.html) for more information.

## Example Usage

```terraform
resource "aws_customerprofiles_domain" "example" {
  domain_name = "example"
}

resource "aws_customerprofiles_profile" "example" {
  domain_name = aws_customerprofiles_domain.example.domain_name
}
```

## Argument Reference

The following arguments are required:

* `domain_name` - The name of your Customer Profile domain. It must be unique for your AWS account.

The following arguments are optional:

* `account_number` - A unique account number that you have given to the customer.
* `additional_information` - Any additional information relevant to the customer’s profile.
* `address` - A block that specifies a generic address associated with the customer that is not mailing, shipping, or billing. [Documented below](#address).
* `attributes` - A key value pair of attributes of a customer profile.
* `billing_address` - A block that specifies the customer’s billing address. [Documented below](#address).
* `birth_date` - The customer’s birth date.
* `business_email_address` - The customer’s business email address.
* `business_name` - The name of the customer’s business.
* `business_phone_number` - The customer’s business phone number.
* `email_address` - The customer’s email address, which has not been specified as a personal or business address.
* `first_name` - The customer’s first name.
* `gender_string` - The gender with which the customer identifies.
* `home_phone_number` - The customer’s home phone number.
* `last_name` - The customer’s last name.
* `mailing_address` - A block that specifies the customer’s mailing address. [Documented below](#address).
* `middle_name` - The customer’s middle name.
* `mobile_phone_number` - The customer’s mobile phone number.
* `party_type_string` - The type of profile used to describe the customer.
* `personal_email_address` - The customer’s personal email address.
* `phone_number` - The customer’s phone number, which has not been specified as a mobile, home, or business number.
* `shipping_address` - A block that specifies the customer’s shipping address. [Documented below](#address).

### `address`

The `address` configuration block supports the following attributes:

* `address_1` - The first line of a customer address.
* `address_2` - The second line of a customer address.
* `address_3` - The third line of a customer address.
* `address_4` - The fourth line of a customer address.
* `city` - The city in which a customer lives.
* `country` - The country in which a customer lives.
* `county` - The county in which a customer lives.
* `postal_code` - The postal code of a customer address.
* `province` - The province in which a customer lives.
* `state` - The state in which a customer lives.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - The identifier of the Customer Profiles Profile.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `30m`)
* `update` - (Default `30m`)
* `delete` - (Default `30m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Amazon Customer Profiles Profile using the resource `id`. For example:

```terraform
import {
  to = aws_customerprofiles_profile.example
  id = "domain-name/5f2f473dfbe841eb8d05cfc2a4c926df"
}
```

Using `terraform import`, import Amazon Customer Profiles Profile using the resource `id`. For example:

```console
% terraform import aws_customerprofiles_profile.example domain-name/5f2f473dfbe841eb8d05cfc2a4c926df
```
