---
subcategory: "RDS (Relational Database)"
layout: "aws"
page_title: "AWS: aws_rds_certificate"
description: |-
  Information about an RDS Certificate.
---

# Data Source: aws_rds_certificate

Information about an RDS Certificate.

## Example Usage

```terraform
data "aws_rds_certificate" "example" {
  latest_valid_till = true
}
```

## Argument Reference

This data source supports the following arguments:

* `id` - (Optional) Certificate identifier. For example, `rds-ca-2019`.
* `default_for_new_launches` - (Optional) When enabled, returns the default certificate for new RDS instances.
* `latest_valid_till` - (Optional) When enabled, returns the certificate with the latest `ValidTill`.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `arn` - ARN of the certificate.
* `certificate_type` - Type of certificate. For example, `CA`.
* `customer_override` - Boolean whether there is an override for the default certificate identifier.
* `customer_override_valid_till` - If there is an override for the default certificate identifier, when the override expires.
* `thumbprint` - Thumbprint of the certificate.
* `valid_from` - [RFC3339 format](https://tools.ietf.org/html/rfc3339#section-5.8) of certificate starting validity date.
* `valid_till` - [RFC3339 format](https://tools.ietf.org/html/rfc3339#section-5.8) of certificate ending validity date.
