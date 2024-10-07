---
subcategory: "Route 53 Profiles"
layout: "aws"
page_title: "AWS: aws_route53profiles_association"
description: |-
  Terraform resource for managing an AWS Route 53 Profiles Association.
---

# Resource: aws_route53profiles_association

Terraform resource for managing an AWS Route 53 Profiles Association.

## Example Usage

### Basic Usage

```terraform
resource "aws_route53profiles_profile" "example" {
  name = "example"
}

resource "aws_vpc" "example" {
  cidr = "10.0.0.0/16"
}

resource "aws_route53profiles_association" "example" {
  name        = "example"
  profile_id  = aws_route53profiles_profile.example.id
  resource_id = aws_vpc.example.id
}
```

## Argument Reference

The following arguments are required:

* `name` - (Required) Name of the Profile Association.

* `profile_id` - (Required) ID of the profile associated with the VPC.

* `resource_id` - (Required) Resource ID of the VPC the profile to be associated with.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - ID of the Profile Association.
* `name` - Name of the Profile Association.
* `status` - Status of the Profile Association. Valid values [AWS docs](https://docs.aws.amazon.com/Route53/latest/APIReference/API_route53profiles_Profile.html)
* `status_message` - Status message of the Profile Association.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `30m`)
* `delete` - (Default `30m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Route 53 Profiles Association using the `id`. For example:

```terraform
import {
  to = aws_route53profiles_association.example
  id = "rpa-id-12345678"
}
```

Using `terraform import`, import Route 53 Profiles Association using the `example_id_arg`. For example:

```console
% terraform import aws_route53profiles_association.example rpa-id-12345678
```
