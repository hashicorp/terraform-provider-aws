---
subcategory: "Route 53 Resolver"
layout: "aws"
page_title: "AWS: aws_route53_resolver_firewall_domain_list"
description: |-
    Retrieves the specified firewall domain list.
---

# Data Source: aws_route53_resolver_firewall_domain_list

`aws_route53_resolver_firewall_domain_list` Retrieves the specified firewall domain list.

This data source allows to retrieve details about a specific a Route 53 Resolver DNS Firewall domain list.

## Example Usage

The following example shows how to get a firewall domain list from its ID.

```terraform
data "aws_route53_resolver_firewall_domain_list" "example" {
  firewall_domain_list_id = "rslvr-fdl-example"
}
```

## Argument Reference

This data source supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `firewall_domain_list_id` - (Required) The ID of the domain list.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `arn` - The Amazon Resource Name (ARN) of the firewall domain list.
* `creation_time` - The date and time that the domain list was created, in Unix time format and Coordinated Universal Time (UTC).
* `creator_request_id` - A unique string defined by you to identify the request.
* `domain_count` - The number of domain names that are specified in the domain list.
* `name` - The name of the domain list.
* `managed_owner_name` - The owner of the list, used only for lists that are not managed by you.
* `modification_time` - The date and time that the domain list was last modified, in Unix time format and Coordinated Universal Time (UTC).
* `status` - The status of the domain list.
* `status_message` - Additional information about the status of the list, if available.
