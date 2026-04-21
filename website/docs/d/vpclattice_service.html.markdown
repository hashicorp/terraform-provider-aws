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
  name = "example"
}
```

## Argument Reference

This data source supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `name` - (Optional) Service name.
* `service_identifier` - (Optional) ID or Amazon Resource Name (ARN) of the service.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `arn` - ARN of the service.
* `auth_type` - Type of IAM policy. Either `NONE` or `AWS_IAM`.
* `certificate_arn` - Amazon Resource Name (ARN) of the certificate.
* `custom_domain_name` - Custom domain name of the service.
* `dns_entry` - List of objects with DNS names.
    * `domain_name` - DNS name for the service.
    * `hosted_zone_id` - Hosted zone ID where the DNS name is registered.
* `id` - Unique identifier for the service.
* `status` - Status of the service.
* `tags` - List of tags associated with the service.
