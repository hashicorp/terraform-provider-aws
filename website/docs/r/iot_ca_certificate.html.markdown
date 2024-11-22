---
subcategory: "IoT Core"
layout: "aws"
page_title: "AWS: aws_iot_ca_certificate"
description: |-
    Creates and manages an AWS IoT CA Certificate.
---

# Resource: aws_iot_ca_certificate

Creates and manages an AWS IoT CA Certificate.

## Example Usage

```terraform
resource "tls_self_signed_cert" "ca" {
  private_key_pem = tls_private_key.ca.private_key_pem
  subject {
    common_name  = "example.com"
    organization = "ACME Examples, Inc"
  }
  validity_period_hours = 12
  allowed_uses = [
    "key_encipherment",
    "digital_signature",
    "server_auth",
  ]
  is_ca_certificate = true
}

resource "tls_private_key" "ca" {
  algorithm = "RSA"
}

resource "tls_cert_request" "verification" {
  private_key_pem = tls_private_key.verification.private_key_pem
  subject {
    common_name = data.aws_iot_registration_code.example.registration_code
  }
}

resource "tls_private_key" "verification" {
  algorithm = "RSA"
}

resource "tls_locally_signed_cert" "verification" {
  cert_request_pem      = tls_cert_request.verification.cert_request_pem
  ca_private_key_pem    = tls_private_key.ca.private_key_pem
  ca_cert_pem           = tls_self_signed_cert.ca.cert_pem
  validity_period_hours = 12
  allowed_uses = [
    "key_encipherment",
    "digital_signature",
    "server_auth",
  ]
}

resource "aws_iot_ca_certificate" "example" {
  active                       = true
  ca_certificate_pem           = tls_self_signed_cert.ca.cert_pem
  verification_certificate_pem = tls_locally_signed_cert.verification.cert_pem
  allow_auto_registration      = true
}

data "aws_iot_registration_code" "example" {}
```

## Argument Reference

* `active` - (Required)  Boolean flag to indicate if the certificate should be active for device authentication.
* `allow_auto_registration` - (Required)  Boolean flag to indicate if the certificate should be active for device regisration.
* `ca_certificate_pem` - (Required)  PEM encoded CA certificate.
* `certificate_mode` - (Optional)  The certificate mode in which the CA will be registered. Valida values: `DEFAULT` and `SNI_ONLY`. Default: `DEFAULT`.
* `registration_config` - (Optional) Information about the registration configuration. See below.
* `tags` - (Optional) A map of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.
* `verification_certificate_pem` - (Optional) PEM encoded verification certificate containing the common name of a registration code. Review
  [CreateVerificationCSR](https://docs.aws.amazon.com/iot/latest/developerguide/register-CA-cert.html). Reuired if `certificate_mode` is `DEFAULT`.

### registration_config

* `role_arn` - (Optional) The ARN of the role.
* `template_body` - (Optional) The template body.
* `template_name` - (Optional) The name of the provisioning template.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - The internal ID assigned to this CA certificate.
* `arn` - The ARN of the created CA certificate.
* `customer_version` - The customer version of the CA certificate.
* `generation_id` - The generation ID of the CA certificate.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).
* `validity` - When the CA certificate is valid.
    * `not_after` - The certificate is not valid after this date.
    * `not_before` - The certificate is not valid before this date.
