---
subcategory: "VPC Lattice"
layout: "aws"
page_title: "AWS: aws_vpclattice_service"
description: |-
  Terraform data source for managing an AWS VPC Lattice Service.
---

# Data Source: aws_vpclattice_service

Terraform data source for managing an AWS VPC Lattice Service.

## Example Usage

### Basic Usage

```terraform
data "aws_vpclattice_service" "example" {
}
```

## Argument Reference

The following arguments are required:

* `service_identifier` - (Required) ID or Amazon Resource Name (ARN) of the service network

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `arn` - ARN of the service.
* `auth_type` - Type of IAM policy. Either `NONE` or `AWS_IAM`.
* `certificate_arn` - Amazon Resource Name (ARN) of the certificate.
* `custom_domain_name` - Custom domain name of the service.
* `dns_entry` - DNS name of the service.
* `id` - Unique identifier for the service.
* `status` - Status of the service.
* `tags` - List of tags associated with the service.
