---
subcategory: "Route 53"
layout: "aws"
page_title: "AWS: aws_route53_traffic_policy_document"
description: |-
    Generates an Route53 traffic policy document in JSON format
---

# Data Source: aws_route53_traffic_policy_document

Generates an Route53 traffic policy document in JSON format for use with resources that expect policy documents such as [`aws_route53_traffic_policy`](/docs/providers/aws/r/route53_traffic_policy.html).

## Example Usage

### Basic Example

```terraform
data "aws_region" "current" {}

data "aws_route53_traffic_policy_document" "example" {
  record_type = "A"
  start_rule  = "site_switch"

  endpoint {
    id    = "my_elb"
    type  = "elastic-load-balancer"
    value = "elb-111111.${data.aws_region.current.name}.elb.amazonaws.com"
  }
  endpoint {
    id     = "site_down_banner"
    type   = "s3-website"
    region = data.aws_region.current.name
    value  = "www.example.com"
  }

  rule {
    id   = "site_switch"
    type = "failover"

    primary {
      endpoint_reference = "my_elb"
    }
    secondary {
      endpoint_reference = "site_down_banner"
    }
  }
}

resource "aws_route53_traffic_policy" "example" {
  name     = "example"
  comment  = "example comment"
  document = data.aws_route53_traffic_policy_document.example.json
}
```

### Complex Example

The following example showcases the use of nested rules within the traffic policy document and introduces the `geoproximity` rule type.

```terraform
data "aws_route53_traffic_policy_document" "example" {
  record_type = "A"
  start_rule  = "geoproximity_rule"

  # NA Region endpoints
  endpoint {
    id    = "na_endpoint_a"
    type  = "elastic-load-balancer"
    value = "elb-111111.us-west-1.elb.amazonaws.com"
  }

  endpoint {
    id    = "na_endpoint_b"
    type  = "elastic-load-balancer"
    value = "elb-222222.us-west-1.elb.amazonaws.com"
  }

  # EU Region endpoint
  endpoint {
    id    = "eu_endpoint"
    type  = "elastic-load-balancer"
    value = "elb-333333.eu-west-1.elb.amazonaws.com"
  }

  # AP Region endpoint
  endpoint {
    id    = "ap_endpoint"
    type  = "elastic-load-balancer"
    value = "elb-444444.ap-northeast-2.elb.amazonaws.com"
  }

  rule {
    id   = "na_rule"
    type = "failover"

    primary {
      endpoint_reference = "na_endpoint_a"
    }

    secondary {
      endpoint_reference = "na_endpoint_b"
    }

  }

  rule {
    id   = "geoproximity_rule"
    type = "geoproximity"

    geo_proximity_location {
      region                 = "aws:route53:us-west-1"
      bias                   = 10
      evaluate_target_health = true
      rule_reference         = "na_rule"
    }

    geo_proximity_location {
      region                 = "aws:route53:eu-west-1"
      bias                   = 10
      evaluate_target_health = true
      endpoint_reference     = "eu_endpoint"
    }

    geo_proximity_location {
      region                 = "aws:route53:ap-northeast-2"
      bias                   = 0
      evaluate_target_health = true
      endpoint_reference     = "ap_endpoint"
    }
  }

}

resource "aws_route53_traffic_policy" "example" {
  name     = "example"
  comment  = "example comment"
  document = data.aws_route53_traffic_policy_document.example.json
}
```

## Argument Reference

The following arguments are optional:

* `endpoint` (Optional) - Configuration block for the definitions of the endpoints that you want to use in this traffic policy. See below
* `record_type` (Optional) - DNS type of all of the resource record sets that Amazon Route 53 will create based on this traffic policy.
* `rule` (Optional) - Configuration block for definitions of the rules that you want to use in this traffic policy. See below
* `start_endpoint` (Optional) - An endpoint to be as the starting point for the traffic policy.
* `start_rule` (Optional) - A rule to be as the starting point for the traffic policy.
* `version` (Optional) - Version of the traffic policy format.

### `endpoint`

* `id` - (Required) ID of an endpoint you want to assign.
* `type` - (Optional) Type of the endpoint. Valid values are `value`, `cloudfront`, `elastic-load-balancer`, `s3-website`, `application-load-balancer`, `network-load-balancer` and `elastic-beanstalk`
* `region` - (Optional) To route traffic to an Amazon S3 bucket that is configured as a website endpoint, specify the region in which you created the bucket for `region`.
* `value` - (Optional) Value of the `type`.

### `rule`

* `id` - (Required) ID of a rule you want to assign.
* `type` - (Optional) Type of the rule.
* `primary` - (Optional) Configuration block for the settings for the rule or endpoint that you want to route traffic to whenever the corresponding resources are available. Only valid for `failover` type. See below
* `secondary` - (Optional) Configuration block for the rule or endpoint that you want to route traffic to whenever the primary resources are not available. Only valid for `failover` type. See below
* `location` - (Optional) Configuration block for when you add a geolocation rule, you configure your traffic policy to route your traffic based on the geographic location of your users.  Only valid for `geo` type. See below
* `geo_proximity_location` - (Optional) Configuration block for when you add a geoproximity rule, you configure Amazon Route 53 to route traffic to your resources based on the geographic location of your resources. Only valid for `geoproximity` type. See below
* `regions` - (Optional) Configuration block for when you add a latency rule, you configure your traffic policy to route your traffic based on the latency (the time delay) between your users and the AWS regions where you've created AWS resources such as ELB load balancers and Amazon S3 buckets. Only valid for `latency` type. See below
* `items` - (Optional) Configuration block for when you add a multivalue answer rule, you configure your traffic policy to route traffic approximately randomly to your healthy resources.  Only valid for `multivalue` type. See below

### `primary` and `secondary`

* `endpoint_reference` - (Optional) References to an endpoint.
* `evaluate_target_health` - (Optional) Indicates whether you want Amazon Route 53 to evaluate the health of the endpoint and route traffic only to healthy endpoints.
* `health_check` - (Optional) If you want to associate a health check with the endpoint or rule.
* `rule_reference` - (Optional) References to a rule.

### `location`

* `continent` - (Optional) Value of a continent.
* `country` - (Optional) Value of a country.
* `endpoint_reference` - (Optional) References to an endpoint.
* `evaluate_target_health` - (Optional) Indicates whether you want Amazon Route 53 to evaluate the health of the endpoint and route traffic only to healthy endpoints.
* `health_check` - (Optional) If you want to associate a health check with the endpoint or rule.
* `is_default` - (Optional) Indicates whether this set of values represents the default location.
* `rule_reference` - (Optional) References to a rule.
* `subdivision` - (Optional) Value of a subdivision.

### `geo_proximity_location`

* `bias` - (Optional) Specify a value for `bias` if you want to route more traffic to an endpoint from nearby endpoints (positive values) or route less traffic to an endpoint (negative values).
* `endpoint_reference` - (Optional) References to an endpoint.
* `evaluate_target_health` - (Optional) Indicates whether you want Amazon Route 53 to evaluate the health of the endpoint and route traffic only to healthy endpoints.
* `health_check` - (Optional) If you want to associate a health check with the endpoint or rule.
* `latitude` - (Optional) Represents the location south (negative) or north (positive) of the equator. Valid values are -90 degrees to 90 degrees.
* `longitude` - (Optional) Represents the location west (negative) or east (positive) of the prime meridian. Valid values are -180 degrees to 180 degrees.
* `region` - (Optional) If your endpoint is an AWS resource, specify the AWS Region that you created the resource in.
* `rule_reference` - (Optional) References to a rule.

### `region`

* `endpoint_reference` - (Optional) References to an endpoint.
* `evaluate_target_health` - (Optional) Indicates whether you want Amazon Route 53 to evaluate the health of the endpoint and route traffic only to healthy endpoints.
* `health_check` - (Optional) If you want to associate a health check with the endpoint or rule.
* `region` - (Optional) Region code for the AWS Region that you created the resource in.
* `rule_reference` - (Optional) References to a rule.

### `item`

* `endpoint_reference` - (Optional) References to an endpoint.
* `health_check` - (Optional) If you want to associate a health check with the endpoint or rule.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `json` - Standard JSON policy document rendered based on the arguments above.
