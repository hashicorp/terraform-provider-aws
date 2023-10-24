---
subcategory: "VPC Lattice"
layout: "aws"
page_title: "AWS: aws_vpclattice_service_network"
description: |-
  Terraform resource for managing an AWS VPC Lattice Service Network.
---

# Resource: aws_vpclattice_service_network

Terraform resource for managing an AWS VPC Lattice Service Network.

## Example Usage

### Basic Usage

```terraform
resource "aws_vpclattice_service_network" "example" {
  name      = "example"
  auth_type = "AWS_IAM"
}
```

## Argument Reference

The following arguments are required:

* `name` - (Required) Name of the service network

The following arguments are optional:

* `auth_type` - (Optional) Type of IAM policy. Either `NONE` or `AWS_IAM`.
* `tags` - (Optional) Key-value mapping of resource tags. If configured with a provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the Service Network.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block).

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import VPC Lattice Service Network using the `id`. For example:

```terraform
import {
  to = aws_vpclattice_service_network.example
  id = "sn-0158f91c1e3358dba"
}
```

Using `terraform import`, import VPC Lattice Service Network using the `id`. For example:

```console
% terraform import aws_vpclattice_service_network.example sn-0158f91c1e3358dba
```
