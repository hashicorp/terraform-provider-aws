---
layout: "aws"
page_title: "AWS: aws_route53_domain"
sidebar_current: "docs-aws-resource-route53-domain"
description: |-
  Manages a Route53 Domain
---

# Resource: aws_route53_domain

Manages a Route53 Domain.

## Example Usage

```hcl
resource "aws_route53_domain" "dev" {
  name = "dev.example.com"

  tags = {
    Environment = "dev"
  }
}
```

## Argument Reference

The following arguments are supported:

* `domain_name` - (Required) The name of a domain.
* `admin_contact` - (Optional) Provides details about the domain administrative contact.
* `admin_privacy` - (Optional) Specifies whether contact information is concealed from WHOIS queries. If the value is true, WHOIS ("who is") queries return contact information either for Amazon Registrar (for .com, .net, and .org domains) or for our registrar associate, Gandi (for all other TLDs). If the value is false, WHOIS queries return the information that you entered for the admin contact.
* `auto_renew` - (Optional) Specifies whether the domain registration is set to renew automatically.
* `name_servers` - (Optional) The nameservers for the domain.
* `registrant_contact` - (Optional) Provides details about the domain registrant.
* `registrant_privacy` - (Optional) Specifies whether contact information is concealed from WHOIS queries. If the value is true, WHOIS ("who is") queries return contact information either for Amazon Registrar (for .com, .net, and .org domains) or for our registrar associate, Gandi (for all other TLDs). If the value is false, WHOIS queries return the information that you entered for the registrant contact (domain owner).
* `tech_contact` - (Optional) Provides details about the domain technical contact.
* `tech_privacy` - (Optional) Specifies whether contact information is concealed from WHOIS queries. If the value is true, WHOIS ("who is") queries return contact information either for Amazon Registrar (for .com, .net, and .org domains) or for our registrar associate, Gandi (for all other TLDs). If the value is false, WHOIS queries return the information that you entered for the technical contact.
* `tags` - (Optional) A mapping of tags to assign to the domain.

`admin_contact`, `registrant_contact` and `tech_contact` support the following arguments:

* `address_line_1` - (Optional) First line of the contact's address.
* `address_line_2` - (Optional) Second line of contact's address, if any.
* `city` - (Optional) The city of the contact's address.
* `contact_type` - (Optional) Indicates whether the contact is a person, company, association, or public organization. One of `PERSON`, `COMPANY`, `ASSOCIATION`, `PUBLIC_BODY`, `RESELLER`
* `country_code` - (Optional) Code for the country of the contact's address.
* `email` - (Optional) Email address of the contact.
* `fax` - (Optional) Fax number of the contact. Phone number must be specified in the format "+[country dialing code].[number including any area code]". For example, a US phone number might appear as "+1.1234567890".
* `first_name` - (Optional) First name of contact.
* `last_name` - (Optional) Last name of contact.
* `organization_name` - (Optional) Name of the organization for contact types other than `PERSON`.
* `phone_number` - (Optional) The phone number of the contact. Phone number must be specified in the format "+[country dialing code].[number including any area code]". For example, a US phone number might appear as "+1.1234567890".
* `state` - (Optional) The state or province of the contact's city.
* `zip_code` - (Optional) The zip or postal code of the contact's address.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `abuse_contact_email` - Email address to contact to report incorrect contact information for a domain, to report that the domain is being used to send spam, to report that someone is cybersquatting on a domain name, or report some other type of abuse.
* `abuse_contact_phone` - Phone number for reporting abuse.
* `creation_date` - The date when the domain was created as found in the response to a WHOIS query. The date and time is in Unix time format and Coordinated Universal time (UTC).
* `expiration_date` - The date when the registration for the domain is set to expire. The date and time is in Unix time format and Coordinated Universal time (UTC).
* `registrar_name` - Name of the registrar of the domain as identified in the registry. Domains with a .com, .net, or .org TLD are registered by Amazon Registrar. All other domains are registered by our registrar associate, Gandi. The value for domains that are registered by Gandi is "GANDI SAS".
* `registrar_url` - Web address of the registrar.
* `status_list` - An array of domain name status codes, also known as Extensible Provisioning Protocol (EPP) status codes.
* `updated_date` - The last updated date of the domain as found in the response to a WHOIS query. The date and time is in Unix time format and Coordinated Universal time (UTC).
* `whois_server` - The fully qualified name of the WHOIS server that can answer the WHOIS query for the domain.

## Import

Route53 Domains can be imported using the `domain id`, e.g.

```
$ terraform import aws_route53_domain.mydomain example.com
```
