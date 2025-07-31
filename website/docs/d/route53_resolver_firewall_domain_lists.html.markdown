---
subcategory: "Route 53 Resolver"
layout: "aws"
page_title: "AWS: aws_route53_resolver_firewall_domain_lists"
description: |-
    Provides details about a set of Route53 Resolver DNS Firewall domain lists
---

# Data Source: aws_route53_resolver_firewall_domain_lists

`aws_route53_resolver_firewall_domain_lists` provides details about a set of Route53 Resolver DNS Firewall domain lists.

## Example Usage

### Basic usage

```terraform
data "aws_route53_resolver_firewall_domain_lists" "example" {}

# Find AWS managed malware domain list by filtering the results
locals {
  malware_list = one([
    for list in data.aws_route53_resolver_firewall_domain_lists.example.firewall_domain_lists :
    list if list.name == "AWSManagedDomainsMalwareDomainList"
  ])
}

resource "aws_route53_resolver_firewall_rule" "example" {
  name                    = "block-malware-domains"
  action                  = "BLOCK"
  block_response          = "NXDOMAIN"
  firewall_domain_list_id = local.malware_list.id
  firewall_rule_group_id  = aws_route53_resolver_firewall_rule_group.example.id
  priority                = 100
}

# Or access additional details
output "malware_list_arn" {
  value = local.malware_list.arn
}
```

## Argument Reference

This data source supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `id` - AWS Region.
* `firewall_domain_lists` - List of firewall domain lists with detailed information. Each element contains:
  * `arn` - The Amazon Resource Name (ARN) of the firewall domain list.
  * `creator_request_id` - A unique string defined by you to identify the request.
  * `id` - The ID of the firewall domain list.
  * `managed_owner_name` - The owner of the firewall domain list. For AWS managed domain lists, this is `Route 53 Resolver DNS Firewall`.
  * `name` - The name of the firewall domain list.