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
  tags = {
    Environment = "dev"
  }
}
```

## Argument Reference

This resource supports the following arguments:

* `name` - (Required) Name of the Profile.
* `tags` - (Optional) Map of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the Profile.
* `id` - ID of the Profile.
* `name` - Name of the Profile.
* `share_status` - Share status of the Profile.
* `status` - Status of the Profile.
* `status_message` - Status message of the Profile.
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
