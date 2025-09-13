---
subcategory: "Route 53 Resolver"
layout: "aws"
page_title: "AWS: aws_route53_resolver_firewall_rule_group_association"
description: |-
  Provides a Route 53 Resolver DNS Firewall rule group association resource.
---

# Resource: aws_route53_resolver_firewall_rule_group_association

Provides a Route 53 Resolver DNS Firewall rule group association resource.

## Example Usage

```terraform
resource "aws_route53_resolver_firewall_rule_group" "example" {
  name = "example"
}

resource "aws_route53_resolver_firewall_rule_group_association" "example" {
  name                   = "example"
  firewall_rule_group_id = aws_route53_resolver_firewall_rule_group.example.id
  priority               = 100
  vpc_id                 = aws_vpc.example.id
}
```

## Argument Reference

This resource supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `name` - (Required) A name that lets you identify the rule group association, to manage and use it.
* `firewall_rule_group_id` - (Required) The unique identifier of the firewall rule group.
* `mutation_protection` - (Optional) If enabled, this setting disallows modification or removal of the association, to help prevent against accidentally altering DNS firewall protections. Valid values: `ENABLED`, `DISABLED`.
* `priority` - (Required) The setting that determines the processing order of the rule group among the rule groups that you associate with the specified VPC. DNS Firewall filters VPC traffic starting from the rule group with the lowest numeric priority setting.
* `vpc_id` - (Required) The unique identifier of the VPC that you want to associate with the rule group.
* `tags` - (Optional) Key-value map of resource tags. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - The ARN (Amazon Resource Name) of the firewall rule group association.
* `id` - The identifier for the association.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Route 53 Resolver DNS Firewall rule group associations using the Route 53 Resolver DNS Firewall rule group association ID. For example:

```terraform
import {
  to = aws_route53_resolver_firewall_rule_group_association.example
  id = "rslvr-frgassoc-0123456789abcdef"
}
```

Using `terraform import`, import Route 53 Resolver DNS Firewall rule group associations using the Route 53 Resolver DNS Firewall rule group association ID. For example:

```console
% terraform import aws_route53_resolver_firewall_rule_group_association.example rslvr-frgassoc-0123456789abcdef
```
