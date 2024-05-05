---
subcategory: "Route 53"
layout: "aws"
page_title: "AWS: aws_route53_record"
description: |-
  Provides a Route53 record resource.
---

# Resource: aws_route53_record

Provides a Route53 record resource.

## Example Usage

### Simple routing policy

```terraform
resource "aws_route53_record" "www" {
  zone_id = aws_route53_zone.primary.zone_id
  name    = "www.example.com"
  type    = "A"
  ttl     = 300
  records = [aws_eip.lb.public_ip]
}
```

### Weighted routing policy

Other routing policies are configured similarly. See [Amazon Route 53 Developer Guide](https://docs.aws.amazon.com/Route53/latest/DeveloperGuide/routing-policy.html) for details.

```terraform
resource "aws_route53_record" "www-dev" {
  zone_id = aws_route53_zone.primary.zone_id
  name    = "www"
  type    = "CNAME"
  ttl     = 5

  weighted_routing_policy {
    weight = 10
  }

  set_identifier = "dev"
  records        = ["dev.example.com"]
}

resource "aws_route53_record" "www-live" {
  zone_id = aws_route53_zone.primary.zone_id
  name    = "www"
  type    = "CNAME"
  ttl     = 5

  weighted_routing_policy {
    weight = 90
  }

  set_identifier = "live"
  records        = ["live.example.com"]
}
```

### Geoproximity routing policy

```terraform
resource "aws_route53_record" "www" {
  zone_id = aws_route53_zone.primary.zone_id
  name    = "www.example.com"
  type    = "CNAME"
  ttl     = 300
  geoproximity_routing_policy {
    coordinates {
      latitude  = "49.22"
      longitude = "-74.01"
    }
  }
  set_identifier = "dev"
  records        = ["dev.example.com"]
}
```

### Alias record

See [related part of Amazon Route 53 Developer Guide](https://docs.aws.amazon.com/Route53/latest/DeveloperGuide/resource-record-sets-choosing-alias-non-alias.html)
to understand differences between alias and non-alias records.

TTL for all alias records is [60 seconds](https://aws.amazon.com/route53/faqs/#dns_failover_do_i_need_to_adjust),
you cannot change this, therefore `ttl` has to be omitted in alias records.

```terraform
resource "aws_elb" "main" {
  name               = "foobar-terraform-elb"
  availability_zones = ["us-east-1c"]

  listener {
    instance_port     = 80
    instance_protocol = "http"
    lb_port           = 80
    lb_protocol       = "http"
  }
}

resource "aws_route53_record" "www" {
  zone_id = aws_route53_zone.primary.zone_id
  name    = "example.com"
  type    = "A"

  alias {
    name                   = aws_elb.main.dns_name
    zone_id                = aws_elb.main.zone_id
    evaluate_target_health = true
  }
}
```

### NS and SOA Record Management

When creating Route 53 zones, the `NS` and `SOA` records for the zone are automatically created. Enabling the `allow_overwrite` argument will allow managing these records in a single Terraform run without the requirement for `terraform import`.

```terraform
resource "aws_route53_zone" "example" {
  name = "test.example.com"
}

resource "aws_route53_record" "example" {
  allow_overwrite = true
  name            = "test.example.com"
  ttl             = 172800
  type            = "NS"
  zone_id         = aws_route53_zone.example.zone_id

  records = [
    aws_route53_zone.example.name_servers[0],
    aws_route53_zone.example.name_servers[1],
    aws_route53_zone.example.name_servers[2],
    aws_route53_zone.example.name_servers[3],
  ]
}
```

## Argument Reference

This resource supports the following arguments:

* `zone_id` - (Required) The ID of the hosted zone to contain this record.
* `name` - (Required) The name of the record.
* `type` - (Required) The record type. Valid values are `A`, `AAAA`, `CAA`, `CNAME`, `DS`, `MX`, `NAPTR`, `NS`, `PTR`, `SOA`, `SPF`, `SRV` and `TXT`.
* `ttl` - (Required for non-alias records) The TTL of the record.
* `records` - (Required for non-alias records) A string list of records. To specify a single record value longer than 255 characters such as a TXT record for DKIM, add `\"\"` inside the Terraform configuration string (e.g., `"first255characters\"\"morecharacters"`).
* `set_identifier` - (Optional) Unique identifier to differentiate records with routing policies from one another. Required if using `cidr_routing_policy`, `failover_routing_policy`, `geolocation_routing_policy`,`geoproximity_routing_policy`, `latency_routing_policy`, `multivalue_answer_routing_policy`, or `weighted_routing_policy`.
* `health_check_id` - (Optional) The health check the record should be associated with.
* `alias` - (Optional) An alias block. Conflicts with `ttl` & `records`.
  [Documented below](#alias).
* `cidr_routing_policy` - (Optional) A block indicating a routing policy based on the IP network ranges of requestors. Conflicts with any other routing policy. [Documented below](#cidr-routing-policy).
* `failover_routing_policy` - (Optional) A block indicating the routing behavior when associated health check fails. Conflicts with any other routing policy. [Documented below](#failover-routing-policy).
* `geolocation_routing_policy` - (Optional) A block indicating a routing policy based on the geolocation of the requestor. Conflicts with any other routing policy. [Documented below](#geolocation-routing-policy).
* `geoproximity_routing_policy` - (Optional) A block indicating a routing policy based on the geoproximity of the requestor. Conflicts with any other routing policy. [Documented below](#geoproximity-routing-policy).
* `latency_routing_policy` - (Optional) A block indicating a routing policy based on the latency between the requestor and an AWS region. Conflicts with any other routing policy. [Documented below](#latency-routing-policy).
* `multivalue_answer_routing_policy` - (Optional) Set to `true` to indicate a multivalue answer routing policy. Conflicts with any other routing policy.
* `weighted_routing_policy` - (Optional) A block indicating a weighted routing policy. Conflicts with any other routing policy. [Documented below](#weighted-routing-policy).
* `allow_overwrite` - (Optional) Allow creation of this record in Terraform to overwrite an existing record, if any. This does not affect the ability to update the record in Terraform and does not prevent other resources within Terraform or manual Route 53 changes outside Terraform from overwriting this record. `false` by default. This configuration is not recommended for most environments.

Exactly one of `records` or `alias` must be specified: this determines whether it's an alias record.

### Alias

Alias records support the following:

* `name` - (Required) DNS domain name for a CloudFront distribution, S3 bucket, ELB, or another resource record set in this hosted zone.
* `zone_id` - (Required) Hosted zone ID for a CloudFront distribution, S3 bucket, ELB, or Route 53 hosted zone. See [`resource_elb.zone_id`](/docs/providers/aws/r/elb.html#zone_id) for example.
* `evaluate_target_health` - (Required) Set to `true` if you want Route 53 to determine whether to respond to DNS queries using this resource record set by checking the health of the resource record set. Some resources have special requirements, see [related part of documentation](https://docs.aws.amazon.com/Route53/latest/DeveloperGuide/resource-record-sets-values.html#rrsets-values-alias-evaluate-target-health).

### CIDR Routing Policy

CIDR routing policies support the following:

* `collection_id` - (Required) The CIDR collection ID. See the [`aws_route53_cidr_collection` resource](route53_cidr_collection.html) for more details.
* `location_name` - (Required) The CIDR collection location name. See the [`aws_route53_cidr_location` resource](route53_cidr_location.html) for more details. A `location_name` with an asterisk `"*"` can be used to create a default CIDR record. `collection_id` is still required for default record.

### Failover Routing Policy

Failover routing policies support the following:

* `type` - (Required) `PRIMARY` or `SECONDARY`. A `PRIMARY` record will be served if its healthcheck is passing, otherwise the `SECONDARY` will be served. See http://docs.aws.amazon.com/Route53/latest/DeveloperGuide/dns-failover-configuring-options.html#dns-failover-failover-rrsets

### Geolocation Routing Policy

Geolocation routing policies support the following:

* `continent` - A two-letter continent code. See http://docs.aws.amazon.com/Route53/latest/APIReference/API_GetGeoLocation.html for code details. Either `continent` or `country` must be specified.
* `country` - A two-character country code or `*` to indicate a default resource record set.
* `subdivision` - (Optional) A subdivision code for a country.

### GeoproximityRouting Policy

Geoproximity routing policies support the following:

* `aws_region` - A AWS region where the resource is present.
* `bias` - Route more traffic or less traffic to the resource by specifying a value ranges between -90 to 90. See https://docs.aws.amazon.com/Route53/latest/DeveloperGuide/routing-policy-geoproximity.html for bias details.
* `coordinates` - Specify `latitude` and `longitude` for routing traffic to non-AWS resources.
* `local_zone_group` - A AWS local zone group where the resource is present. See https://docs.aws.amazon.com/local-zones/latest/ug/available-local-zones.html for local zone group list.

### Latency Routing Policy

Latency routing policies support the following:

* `region` - (Required) An AWS region from which to measure latency. See http://docs.aws.amazon.com/Route53/latest/DeveloperGuide/routing-policy.html#routing-policy-latency

### Weighted Routing Policy

Weighted routing policies support the following:

* `weight` - (Required) A numeric value indicating the relative weight of the record. See http://docs.aws.amazon.com/Route53/latest/DeveloperGuide/routing-policy.html#routing-policy-weighted.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `name` - The name of the record.
* `fqdn` - [FQDN](https://en.wikipedia.org/wiki/Fully_qualified_domain_name) built using the zone domain and `name`.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Route53 Records using the ID of the record, record name, record type, and set identifier. For example:

Using the ID of the record, which is the zone identifier, record name, and record type, separated by underscores (`_`):

```terraform
import {
  to = aws_route53_record.myrecord
  id = "Z4KAPRWWNC7JR_dev.example.com_NS"
}
```

If the record also contains a set identifier, append it:

```terraform
import {
  to = aws_route53_record.myrecord
  id = "Z4KAPRWWNC7JR_dev.example.com_NS_dev"
}
```

If the record name is the empty string, it can be omitted:

```terraform
import {
  to = aws_route53_record.myrecord
  id = "Z4KAPRWWNC7JR__NS"
}
```

**Using `terraform import` to import** Route53 Records using the ID of the record, record name, record type, and set identifier. For example:

Using the ID of the record, which is the zone identifier, record name, and record type, separated by underscores (`_`):

```console
% terraform import aws_route53_record.myrecord Z4KAPRWWNC7JR_dev.example.com_NS
```

If the record also contains a set identifier, append it:

```console
% terraform import aws_route53_record.myrecord Z4KAPRWWNC7JR_dev.example.com_NS_dev
```
