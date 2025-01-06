---
subcategory: "VPC Lattice"
layout: "aws"
page_title: "AWS: aws_vpclattice_service"
description: |-
  Terraform resource for managing an AWS VPC Lattice Service.
---

# Resource: aws_vpclattice_service

Terraform resource for managing an AWS VPC Lattice Service.

## Example Usage

### Basic Usage

```terraform
resource "aws_vpclattice_service" "example" {
  name               = "example"
  auth_type          = "AWS_IAM"
  custom_domain_name = "example.com"
}
```

## Argument Reference

The following arguments are required:

* `name` - (Required) Name of the service. The name must be unique within the account. The valid characters are a-z, 0-9, and hyphens (-). You can't use a hyphen as the first or last character, or immediately after another hyphen.Must be between 3 and 40 characters in length.

The following arguments are optional:

* `auth_type` - (Optional) Type of IAM policy. Either `NONE` or `AWS_IAM`.
* `certificate_arn` - (Optional) Amazon Resource Name (ARN) of the certificate.
* `custom_domain_name` - (Optional) Custom domain name of the service.
* `tags` - (Optional) Key-value mapping of resource tags. If configured with a provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the service.
* `dns_entry` - DNS name of the service.
* `id` - Unique identifier for the service.
* `status` - Status of the service.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block).

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `5m`)
* `delete` - (Default `5m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import VPC Lattice Service using the `id`. For example:

```terraform
import {
  to = aws_vpclattice_service.example
  id = "svc-06728e2357ea55f8a"
}
```

Using `terraform import`, import VPC Lattice Service using the `id`. For example:

```console
% terraform import aws_vpclattice_service.example svc-06728e2357ea55f8a
```
