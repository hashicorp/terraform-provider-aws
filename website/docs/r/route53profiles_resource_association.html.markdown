---
subcategory: "Route 53 Profiles"
layout: "aws"
page_title: "AWS: aws_route53profiles_resource_association"
description: |-
  Terraform resource for managing an AWS Route 53 Profiles Resource Association.
---

# Resource: aws_route53profiles_resource_association

Terraform resource for managing an AWS Route 53 Profiles Resource Association.

## Example Usage

### Basic Usage

```terraform
resource "aws_route53profiles_profile" "example" {
  name = "example"
}

resource "aws_vpc" "example" {
  cidr = "10.0.0.0/16"
}

resource "aws_route53_zone" "example" {
  name = "example.com"

  vpc {
    vpc_id = aws_vpc.example.id
  }
}

resource "aws_route53profiles_resource_association" "example" {
  name         = "example"
  profile_id   = aws_route53profiles_profile.example.id
  resource_arn = aws_route53_zone.example.arn
}
```

## Argument Reference

The following arguments are required:

* `name` - (Required) Name of the Profile Resource Association.
* `profile_id` - (Required) ID of the profile associated with the VPC.
* `resource_arn` - (Required) Resource ID of the resource to be associated with the profile.
* `resource_properties` - (Optional) Resource properties for the resource to be associated with the profile.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - ID of the Profile Resource Association.
* `name` - Name of the Profile Resource Association.
* `resource_type` - Type of resource associated with the profile.
* `status` - Status of the Profile Association. Valid values [AWS docs](https://docs.aws.amazon.com/Route53/latest/APIReference/API_route53profiles_Profile.html)
* `status_message` - Status message of the Profile Resource Association.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `30m`)
* `update` - (Default `30m`)
* `delete` - (Default `30m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Route 53 Profiles Resource Association using the `id`. For example:

```terraform
import {
  to = aws_route53profiles_resource_association.example
  id = "rpa-id-12345678"
}
```

Using `terraform import`, import Route 53 Profiles Resource Association using the `example_id_arg`. For example:

```console
% terraform import aws_route53profiles_resource_association.example rpa-id-12345678
```
