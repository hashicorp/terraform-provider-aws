---
subcategory: "Route 53"
layout: "aws"
page_title: "AWS: aws_route53_zone"
description: |-
    Provides details about a specific Route 53 Hosted Zone
---

# Data Source: aws_route53_zone

`aws_route53_zone` provides details about a specific Route 53 Hosted Zone.

This data source allows to find a Hosted Zone ID given Hosted Zone name and certain search criteria.

## Example Usage

The following example shows how to get a Hosted Zone from its name and from this data how to create a Record Set.

```terraform
data "aws_route53_zone" "selected" {
  name         = "test.com."
  private_zone = true
}

resource "aws_route53_record" "www" {
  zone_id = data.aws_route53_zone.selected.zone_id
  name    = "www.${data.aws_route53_zone.selected.name}"
  type    = "A"
  ttl     = "300"
  records = ["10.0.0.1"]
}
```

The following example shows how to get a Hosted Zone from a unique combination of its tags:

```terraform
data "aws_route53_zone" "selected" {
  tags {
    scope    = "local"
    category = "api"
  }
}

output "local_api_zone" {
  value = data.aws_route53_zone.selected.zone_id
}
```

## Argument Reference

This data source supports the following arguments:

* `zone_id` - (Optional) Directly return the Hosted Zone with the specified Zone ID. No further filtering is performed.
* `name` - (Optional) Hosted Zone name of the desired Hosted Zone. If blank, then accept any name, filtering on only `private_zone`, `vpc_id` and `tags`.
* `private_zone` - (Optional) Filter to only private Hosted Zones.
* `vpc_id` - (Optional, string) Filter to private Hosted Zones associated with the specified `vpc_id`.
* `tags` - (Optional) A map of tags, each pair of which must exactly match a pair on the desired Hosted Zone.

The arguments of this data source act as filters for querying the available Hosted Zone.

- The given filter must match exactly one Hosted Zone.
- `zone_id` and `name` are mutually exclusive.
- If you use the `name` argument for a private Hosted Zone, you need to set the `private_zone` argument to `true`.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `arn` - ARN of the Hosted Zone.
* `caller_reference` - Caller Reference of the Hosted Zone.
* `comment` - Comment field of the Hosted Zone.
* `linked_service_principal` - The service that created the Hosted Zone (e.g., `servicediscovery.amazonaws.com`).
* `linked_service_description` - The description provided by the service that created the Hosted Zone (e.g., `arn:aws:servicediscovery:us-east-1:1234567890:namespace/ns-xxxxxxxxxxxxxxxx`).
* `name` - The Hosted Zone name.
* `name_servers` - List of DNS name servers for the Hosted Zone.
* `primary_name_server` - The Route 53 name server that created the SOA record.
* `private_zone` - Indicates whether this is a private hosted zone.
* `resource_record_set_count` - The number of Record Set in the Hosted Zone.
* `tags` - A map of tags assigned to the Hosted Zone.
* `zone_id` - The Hosted Zone identifier.
