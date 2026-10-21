---
subcategory: "Route 53"
layout: "aws"
page_title: "AWS: aws_route53_zones"
description: |-
    Provides a list of Route53 Hosted Zone IDs in a Region
---

# Data Source: aws_route53_zones

This resource can be useful for getting back a list of Route53 Hosted Zone IDs for a Region.

## Example Usage

The following example retrieves a list of all Hosted Zone IDs.

```terraform
data "aws_route53_zones" "all" {}

output "example" {
  value = data.aws_route53_zones.all.ids
}
```

### Look Up Zone IDs By Name

The `zones` attribute includes the name of each hosted zone, so a single read of this data source can resolve many domain names to their zone IDs. This avoids one [`aws_route53_zone`](/docs/providers/aws/d/route53_zone.html) data source per domain, each of which lists all of the hosted zones in the account.

```terraform
data "aws_route53_zones" "all" {}

locals {
  zone_ids = {
    for zone in data.aws_route53_zones.all.zones :
    zone.name => zone.zone_id if !zone.private_zone
  }
}

resource "aws_route53_record" "www" {
  for_each = toset(["example.com", "example.net"])

  zone_id = local.zone_ids[each.value]
  name    = "www.${each.value}"
  type    = "CNAME"
  ttl     = 300
  records = ["example.org"]
}
```

## Argument Reference

This data source does not support any arguments.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `ids` - A list of all the Route53 Hosted Zone IDs found.
* `zones` - A list of all the Route53 Hosted Zones found. See [`zones`](#zones) below.

### `zones`

* `name` - Name of the hosted zone, without a trailing period.
* `private_zone` - Whether this is a private hosted zone.
* `resource_record_set_count` - Number of resource record sets in the hosted zone.
* `zone_id` - Hosted zone ID.
