---
subcategory: "Route 53"
layout: "aws"
page_title: "AWS: aws_route53_records"
description: |-
  Get information about a set of Route 53 resource records.
---

# Data Source: aws_route53_records

Use this data source to get the details of resource records in a Route 53 hosted zone.

## Example Usage

### Basic Usage

Return all records in the zone.

```terraform
data "aws_route53_zone" "selected" {
  name         = "test.com."
  private_zone = true
}

data "aws_route53_records" "example" {
  zone_id = data.aws_route53_zone.selected.zone_id
}
```

### Basic Usage with filter

Return the records that starts with `www`.

```terraform
data "aws_route53_zone" "selected" {
  name         = "test.com."
  private_zone = true
}

data "aws_route53_records" "example" {
  zone_id    = data.aws_route53_zone.selected.zone_id
  name_regex = "^www"
}
```

## Argument Reference

This data source supports the following arguments:

* `name_regex` - (Optional) Regex string to apply to the resource record names returned by AWS.
* `zone_id` - (Required) The ID of the hosted zone that contains the resource record sets that you want to list.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `resource_record_sets` - The resource records sets.
    * `alias_target` -  Information about the AWS resource traffic is routed to.
        * `dns_name` - Target DNS name.
        * `evaluate_target_health` - Whether an alias resource record set inherits the health of the referenced AWS resource.
        * `hosted_zone_id` - Target hosted zone ID.
    * `cidr_routing_config` - Information about the CIDR location traffic is routed to.
        * `collection_id` - The CIDR collection ID.
        * `location_name` - The CIDR collection location name.
    * `failover` - `PRIMARY` or `SECONDARY`.
    * `geolocation` - Information about how Amazon Route 53 responds to DNS queries based on the geographic origin of the query.
        * `continent_code` - The two-letter code for the continent.
        * `country_code` - The two-letter code for a country.
        * `subdivision_code` - The two-letter code for a state of the United States.
    * `geoproximity_location` - Information about how Amazon Route 53 responds to DNS queries based on the geographic origin of the query.
        * `aws_region` - The AWS Region the resource you are directing DNS traffic to, is in.
        * `bias` - The bias increases or decreases the size of the geographic region from which Route 53 routes traffic to a resource.
        * `coordinates` - Contains the longitude and latitude for a geographic region.
            * `latitude` - Latitude.
            * `longitude` - Longitude.
        * `local_zone_group` - An AWS Local Zone Group.
    * `health_check_id` - ID of any applicable health check.
    * `multi_value_answer` - Traffic is routed approximately randomly to multiple resources.
    * `name` - The name of the record.
    * `region` - The Amazon EC2 Region of the resource that this resource record set refers to.
    * `resource_records` - The resource records.
        * `value` - The DNS record value.
    * `set_identifier` - An identifier that differentiates among multiple resource record sets that have the same combination of name and type.
    * `traffic_policy_instance_id` - The ID of any traffic policy instance that Route 53 created this resource record set for.
    * `ttl` - The resource record cache time to live (TTL), in seconds.
    * `type` - The DNS record type.
    * `weight` - Among resource record sets that have the same combination of DNS name and type, a value that determines the proportion of DNS queries that Amazon Route 53 responds to using the current resource record set.
