---
subcategory: "SESv2 (Simple Email V2)"
layout: "aws"
page_title: "AWS: aws_sesv2_email_identity"
description: |-
  Terraform data source for managing an AWS SESv2 (Simple Email V2) Email Identity.
---

# Data Source: aws_sesv2_email_identity

Terraform data source for managing an AWS SESv2 (Simple Email V2) Email Identity.

## Example Usage

### Basic Usage

```terraform
data "aws_sesv2_email_identity" "example" {
  email_identity = "example.com"
}
```

## Argument Reference

This data source supports the following arguments:

* `email_identity` - (Required) Name of the email identity.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `arn` - ARN of the Email Identity.
* `configuration_set_name` - Configuration set associated with the email identity.
* `dkim_signing_attributes` - List of objects that contains at most one element with information about the private key and selector that you want to use to configure DKIM for the identity for Bring Your Own DKIM (BYODKIM) for the identity, or, configures the key length to be used for Easy DKIM.
    * `current_signing_key_length` - [Easy DKIM] The key length of the DKIM key pair in use.
    * `domain_signing_private_key` - [Bring Your Own DKIM] Private key used to generate DKIM signatures.
    * `domain_signing_selector` - [Bring Your Own DKIM] Selector added to the DNS configuration for the domain.
    * `last_key_generation_timestamp` - [Easy DKIM] The last time a key pair was generated for this identity.
    * `next_signing_key_length` - [Easy DKIM] The key length of the future DKIM key pair to be generated. This can be changed at most once per day.
    * `signing_attributes_origin` - String that indicates how DKIM was configured for the identity. `AWS_SES` indicates that DKIM was configured for the identity by using Easy DKIM. `EXTERNAL` indicates that DKIM was configured for the identity by using Bring Your Own DKIM (BYODKIM).
    * `status` - Whether or not Amazon SES has successfully located the DKIM records in the DNS records for the domain. See the [AWS SES API v2 Reference](https://docs.aws.amazon.com/ses/latest/APIReference-V2/API_DkimAttributes.html#SES-Type-DkimAttributes-Status) for supported statuses.
    * `tokens` - If you used Easy DKIM to configure DKIM authentication for the domain, then this object contains a set of unique strings that you use to create a set of CNAME records that you add to the DNS configuration for your domain. When Amazon SES detects these records in the DNS configuration for your domain, the DKIM authentication process is complete. If you configured DKIM authentication for the domain by providing your own public-private key pair, then this object contains the selector for the public key.
* `identity_type` - Email identity type. Valid values: `EMAIL_ADDRESS`, `DOMAIN`.
* `tags` - Key-value mapping of resource tags.
* `verification_status` - Verification status of the identity. The status can be one of the following: `PENDING`, `SUCCESS`, `FAILED`, `TEMPORARY_FAILURE`, and `NOT_STARTED`.
* `verified_for_sending_status` - Whether or not the identity is verified.
