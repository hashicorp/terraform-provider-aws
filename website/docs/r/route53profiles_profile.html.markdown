---
subcategory: "Route 53 Profiles"
layout: "aws"
page_title: "AWS: aws_route53profiles_profile"
description: |-
  Terraform resource for managing an AWS Route 53 Profile.
---

# Resource: aws_route53profiles_profile

Terraform resource for managing an AWS Route 53 Profile.

## Example Usage

### Empty Profile

```terraform
resource "aws_route53profiles_profile" "example" {
  name = "example"
}
```

## Argument Reference

The following arguments are required:

* `name` - (Required) The name of the Profile.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - The ARN of the Profile.
* `id` - The ID of the Profile.
* `name` - The name of the Profile.
* `status` - The status of the Profile. Valid values [AWS docs](https://docs.aws.amazon.com/Route53/latest/APIReference/API_route53profiles_Profile.html)
* `status_message` - The status message of the Profile.
* `share_status` - The share status of the Profile. Valid values [AWS docs](https://docs.aws.amazon.com/Route53/latest/APIReference/API_route53profiles_Profile.html)
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `30m`)
* `read` - (Default `30m`)
* `update` - (Default `30m`)
* `delete` - (Default `30m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Route 53 Profiles Profile using the `example_id_arg`. For example:

```terraform
import {
  to = aws_route53profiles_profile.example
  id = "rp-12345678"
}
```

Using `terraform import`, import Route 53 Profiles Profile using the `example`. For example:

```console
% terraform import aws_route53profiles_profile.example rp-12345678
```
