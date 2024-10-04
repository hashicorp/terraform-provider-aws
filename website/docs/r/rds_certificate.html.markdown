---
subcategory: "RDS (Relational Database)"
layout: "aws"
page_title: "AWS: aws_rds_certificate"
description: |-
  Terraform resource for managing an AWS RDS (Relational Database) Certificate.
---

# Resource: aws_rds_certificate

Provides a resource to override the system-default Secure Sockets Layer/Transport Layer Security (SSL/TLS) certificate for Amazon RDS for new DB instances in the current AWS region.

~> **NOTE:** Removing this Terraform resource removes the override. New DB instances will use the system-default certificate for the current AWS region.

## Example Usage

```terraform
resource "aws_rds_certificate" "example" {
  certificate_identifier = "rds-ca-rsa4096-g1"
}
```

## Argument Reference

The following arguments are required:

* `certificate_identifier` - (Required) Certificate identifier. For example, `rds-ca-rsa4096-g1`. Refer to [AWS RDS (Relational Database) Certificate Identifier](https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/UsingWithRDS.SSL.html#UsingWithRDS.SSL.CertificateIdentifier) for more information.

## Attribute Reference

This resource exports no additional attributes.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import the RDS certificate override. For example:

```terraform
import {
  to = aws_rds_certificate.example
  id = default
}
```

Using `terraform import`, import the default EBS encryption state. For example:

```console
% terraform import aws_rds_certificate.example default
```
