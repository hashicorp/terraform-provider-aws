---
layout: "aws"
page_title: "AWS: aws_lightsail_domain_entry"
sidebar_current: "docs-aws-resource-lightsail-domain-entry"
description: |-
  Provides a Lightsail Domain Entry
---

# aws_lightsail_domain_entry

Creates one of the following entry records associated with the domain:
A record, CNAME record, TXT record, or MX record.
The domain must have been created as lightsail domain.

~> **Note:** Lightsail is currently only supported in a limited number of AWS Regions, please see ["Regions and Availability Zones in Amazon Lightsail"](https://lightsail.aws.amazon.com/ls/docs/overview/article/understanding-regions-and-availability-zones-in-amazon-lightsail) for more details

## Example Usage, creating a new domain

```hcl
resource "aws_lightsail_domain" "example" {
	domain_name = "example.com"
}

resource "aws_lightsail_domain_entry" "access" {
	domain_name = "${aws_lightsail_domain.example.id}"
	domain_entry = {
		name = "access.${aws_lightsail_domain.example.id}"
		target = "172.31.32.33"
		type = "A"
	}
}
```

## Argument Reference

The following arguments are supported:

* `domain_name` - (Required) The name of the Lightsail domain to manage
* `domain_entry` - (Required) Describes a domain recordset entry. Fields documented below. See [DomainEntry API Reference](https://docs.aws.amazon.com/lightsail/2016-11-28/api-reference/API_DomainEntry.html)

**domain_entry** requires the following:

* `is_alias` - (Optional) Specifies whether the domain entry is an alias used by the Lightsail load balancer. You can include an alias (A type) record in your request, which points to a load balancer DNS name and routes traffic to your load balancer.
* `name` - (Required) The full FQDN name of the domain.
* `target` - (Required) The target AWS name server. Be sure to also set `is_alias` to `true` when setting up an `A` record for a Lightsail load balancer.
* `type` - (Required) The type of domain entry (e.g., A, CNAME, SOA or NS).

## Attributes Reference

The following attributes are exported in addition to the arguments listed above:

* `id` - The name of domain entry.
