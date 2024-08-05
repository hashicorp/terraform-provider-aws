---
subcategory: "WAF"
layout: "aws"
page_title: "AWS: aws_wafv2_ip_set"
description: |-
  Provides an AWS WAFv2 IP Set resource.
---

# Resource: aws_wafv2_ip_set

Provides a WAFv2 IP Set Resource

## Example Usage

```terraform
resource "aws_wafv2_ip_set" "example" {
  name               = "example"
  description        = "Example IP set"
  scope              = "REGIONAL"
  ip_address_version = "IPV4"
  addresses          = ["1.2.3.4/32", "5.6.7.8/32"]

  tags = {
    Tag1 = "Value1"
    Tag2 = "Value2"
  }
}
```

## Argument Reference

This resource supports the following arguments:

* `name` - (Required, Forces new resource) A friendly name of the IP set.
* `description` - (Optional) A friendly description of the IP set.
* `scope` - (Required, Forces new resource) Specifies whether this is for an AWS CloudFront distribution or for a regional application. Valid values are `CLOUDFRONT` or `REGIONAL`. To work with CloudFront, you must also specify the Region US East (N. Virginia).
* `ip_address_version` - (Required, Forces new resource) Specify IPV4 or IPV6. Valid values are `IPV4` or `IPV6`.
* `addresses` - (Required) Contains an array of strings that specifies zero or more IP addresses or blocks of IP addresses. All addresses must be specified using Classless Inter-Domain Routing (CIDR) notation. WAF supports all IPv4 and IPv6 CIDR ranges except for `/0`.
* `tags` - (Optional) An array of key:value pairs to associate with the resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - A unique identifier for the IP set.
* `arn` - The Amazon Resource Name (ARN) of the IP set.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import WAFv2 IP Sets using `ID/name/scope`. For example:

```terraform
import {
  to = aws_wafv2_ip_set.example
  id = "a1b2c3d4-d5f6-7777-8888-9999aaaabbbbcccc/example/REGIONAL"
}
```

Using `terraform import`, import WAFv2 IP Sets using `ID/name/scope`. For example:

```console
% terraform import aws_wafv2_ip_set.example a1b2c3d4-d5f6-7777-8888-9999aaaabbbbcccc/example/REGIONAL
```
