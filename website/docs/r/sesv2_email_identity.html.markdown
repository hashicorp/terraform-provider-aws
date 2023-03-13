---
subcategory: "SESv2 (Simple Email V2)"
layout: "aws"
page_title: "AWS: aws_sesv2_email_identity"
description: |-
  Terraform resource for managing an AWS SESv2 (Simple Email V2) Email Identity.
---

# Resource: aws_sesv2_email_identity

Terraform resource for managing an AWS SESv2 (Simple Email V2) Email Identity.

## Example Usage

### Basic Usage

#### Email Address Identity

```terraform
resource "aws_sesv2_email_identity" "example" {
  email_identity = "testing@example.com"
}
```

#### Domain Identity

```terraform
resource "aws_sesv2_email_identity" "example" {
  email_identity = "example.com"
}
```

#### Configuration Set

```terraform
resource "aws_sesv2_configuration_set" "example" {
  configuration_set_name = "example"
}

resource "aws_sesv2_email_identity" "example" {
  email_identity         = "example.com"
  configuration_set_name = aws_sesv2_configuration_set.example.configuration_set_name
}
```

#### DKIM Signing Attributes (BYODKIM)

```terraform
resource "tls_private_key" "example" {
  algorithm = "RSA"
}

resource "aws_sesv2_email_identity" "example" {
  email_identity = "example.com"

  dkim_signing_attributes {
    domain_signing_private_key = base64encode(tls_private_key.example.private_key_pem)
    domain_signing_selector    = "example"
  }
}
```

## Argument Reference

The following arguments are supported:

* `email_identity` - (Required) The email address or domain to verify.
* `configuration_set_name` - (Optional) The configuration set to use by default when sending from this identity. Note that any configuration set defined in the email sending request takes precedence.
* `dkim_signing_attributes` - (Optional) The configuration of the DKIM authentication settings for an email domain identity.

### dkim_signing_attributes

* `domain_signing_private_key` - (Optional) [Bring Your Own DKIM] A private key that's used to generate a DKIM signature. The private key must use 1024 or 2048-bit RSA encryption, and must be encoded using base64 encoding.
* `domain_signing_selector` - (Optional) [Bring Your Own DKIM] A string that's used to identify a public key in the DNS configuration for a domain.
* `next_signing_key_length` - (Optional) [Easy DKIM] The key length of the future DKIM key pair to be generated. This can be changed at most once per day. Valid values: `RSA_1024_BIT`, `RSA_2048_BIT`.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - ARN of the Email Identity.
* `dkim_signing_attributes` - An object that contains information about the private key and selector that you want to use to configure DKIM for the identity for Bring Your Own DKIM (BYODKIM) for the identity, or, configures the key length to be used for Easy DKIM.
    * `current_signing_key_length` - [Easy DKIM] The key length of the DKIM key pair in use.
    * `last_key_generation_timestamp` - [Easy DKIM] The last time a key pair was generated for this identity.
    * `next_signing_key_length` - [Easy DKIM] The key length of the future DKIM key pair to be generated. This can be changed at most once per day.
    * `signing_attributes_origin` - A string that indicates how DKIM was configured for the identity. `AWS_SES` indicates that DKIM was configured for the identity by using Easy DKIM. `EXTERNAL` indicates that DKIM was configured for the identity by using Bring Your Own DKIM (BYODKIM).
    * `status` - Describes whether or not Amazon SES has successfully located the DKIM records in the DNS records for the domain. See the [AWS SES API v2 Reference](https://docs.aws.amazon.com/ses/latest/APIReference-V2/API_DkimAttributes.html#SES-Type-DkimAttributes-Status) for supported statuses.
    * `tokens` - If you used Easy DKIM to configure DKIM authentication for the domain, then this object contains a set of unique strings that you use to create a set of CNAME records that you add to the DNS configuration for your domain. When Amazon SES detects these records in the DNS configuration for your domain, the DKIM authentication process is complete. If you configured DKIM authentication for the domain by providing your own public-private key pair, then this object contains the selector for the public key.
* `identity_type` - The email identity type. Valid values: `EMAIL_ADDRESS`, `DOMAIN`.
* `tags` - (Optional) A map of tags to assign to the service. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.
* `verified_for_sending_status` - Specifies whether or not the identity is verified.

## Import

SESv2 (Simple Email V2) Email Identity can be imported using the `email_identity`, e.g.,

```
$ terraform import aws_sesv2_email_identity.example example.com
```
