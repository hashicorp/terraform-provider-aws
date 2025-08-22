---
subcategory: "IoT Core"
layout: "aws"
page_title: "AWS: aws_iot_registration_code"
description: |-
  Gets a registration code used to register a CA certificate with AWS IoT
---

# Data Source: aws_iot_registration_code

Gets a registration code used to register a CA certificate with AWS IoT.

## Example Usage

```terraform
data "aws_iot_registration_code" "example" {}

resource "tls_private_key" "verification" {
  algorithm = "RSA"
}

resource "tls_cert_request" "verification" {
  key_algorithm   = "RSA"
  private_key_pem = tls_private_key.verification.private_key_pem
  subject {
    common_name = data.aws_iot_registration_code.example.registration_code
  }
}
```

## Argument Reference

This data source supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `registration_code` - The CA certificate registration code.
