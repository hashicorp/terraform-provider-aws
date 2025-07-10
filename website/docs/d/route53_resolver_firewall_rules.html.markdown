---
subcategory: "Route 53 Resolver"
layout: "aws"
page_title: "AWS: aws_route53_resolver_firewall_rules"
description: |-
    Provides details about rules in a specific Route53 Resolver Firewall rule group.
---

# Data Source: aws_route53_resolver_firewall_rules

`aws_route53_resolver_firewall_rules` Provides details about rules in a specific Route53 Resolver Firewall rule group.

## Example Usage

The following example shows how to get Route53 Resolver Firewall rules based on its associated firewall group id.

```terraform
data "aws_route53_resolver_firewall_rules" "example" {
  firewall_rule_group_id = aws_route53_resolver_firewall_rule_group.example.id
}
```

## Argument Reference

This data source supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `firewall_rule_group_id` - (Required) The unique identifier of the firewall rule group that you want to retrieve the rules for.
* `action` - (Optional) The action that DNS Firewall should take on a DNS query when it matches one of the domains in the rule's domain list.
* `priority` - (Optional) The setting that determines the processing order of the rules in a rule group.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `firewall_rules` - List with information about the firewall rules. See details below.

### provisioning_artifact_details

* `block_override_dns_type` - The DNS record's type.
* `block_override_domain` - The custom DNS record to send back in response to the query.
* `block_override_ttl` - The recommended amount of time, in seconds, for the DNS resolver or web browser to cache the provided override record.
* `block_response` - The way that you want DNS Firewall to block the request.
* `creation_time` - The date and time that the rule was created, in Unix time format and Coordinated Universal Time (UTC).
* `creator_request_id` - A unique string defined by you to identify the request.
* `firewall_domain_list_id` - The ID of the domain list that's used in the rule.
* `modification_time` - The date and time that the rule was last modified, in Unix time format and Coordinated Universal Time (UTC).
* `name` - The name of the rule.
