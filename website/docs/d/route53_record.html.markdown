---
subcategory: "Route 53"
layout: "aws"
page_title: "AWS: aws_route53_record"
description: |-
  Terraform data source for managing an AWS Route 53 Record.
---

# Data Source: aws_route53_record

`aws_route53_record` provides details about a records present in Route 53 Hosted Zone.

This data source allows to find a Route53 records given Hosted Zone id and certain search criteria.

## Example Usage

### Basic Usage

```terraform
data "aws_route53_zone" "selected" {
  name         = "test.com."
  private_zone = true
}

data "aws_route53_record" "example" {
  zone_id = data.aws_route53_zone.selected.zone_id
}
```

### Basic Usage with filter

Filters the records that starts with www

```terraform
data "aws_route53_zone" "selected" {
  name         = "test.com."
  private_zone = true
}

data "aws_route53_record" "example" {
  zone_id        = data.aws_route53_zone.selected.zone_id
  filter_record  = "^www"
}
```

## Argument Reference

The argument `filter_record` act as filters for querying the available Route53 records in the Hosted Zone.

The following arguments are required:

* `zone_id` - (Required) Hosted Zone id of the desired Hosted Zone..

The following arguments are optional:

* `filter_record` - (Optional) A regular expression used to filter the records based on certain criteria.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `alias` - The alias name for the record.
* `failover` - Failover configurations.
* `geolocation` - Geolocation configurations
* `geoproximity` - Geoproximity configurations
* `latency_region` - Latency region
* `multivalue_answer` - Multivalue answer enabled or disabled.
* `name` - The name of the record.
* `routing_policy` - Type of the routing policy for the record.
* `set_identifier` - Unique identifier to differentiate records with routing policies from one another
* `ttl` - The TTL of the record.
* `type` - The record type. Example `A`, `AAAA`, `CAA`, `CNAME`, `DS`, `MX`, `NAPTR`, `NS`, `PTR`, `SOA`, `SPF`, `SRV` and `TXT`.
* `values` - List of IP address or DNS record name associated with the route53 record.
* `weight` - The weight of the record.
