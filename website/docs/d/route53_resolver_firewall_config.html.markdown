---
subcategory: "Route 53 Resolver"
layout: "aws"
page_title: "AWS: aws_route53_resolver_firewall_config"
description: |-
    Provides details about a specific a Route 53 Resolver DNS Firewall config.
---

# Data Source: aws_route53_resolver_firewall_config

`aws_route53_resolver_firewall_config` provides details about a specific a Route 53 Resolver DNS Firewall config.

This data source allows to find a details about a specific a Route 53 Resolver DNS Firewall config.

## Example Usage

The following example shows how to get a firewall config using the VPC ID.

```terraform
data "aws_route53_resolver_firewall_config" "example" {
  resource_id = "vpc-exampleid"
}
```

## Argument Reference

* `resource_id` - (Required) The ID of the VPC from Amazon VPC that the configuration is for.

The following attribute is additionally exported:

* `firewall_fail_open` - Determines how DNS Firewall operates during failures, for example when all traffic that is sent to DNS Firewall fails to receive a reply.
* `id` - The ID of the firewall configuration.
* `owner_id` - The Amazon Web Services account ID of the owner of the VPC that this firewall configuration applies to.
