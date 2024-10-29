---
subcategory: "App Runner"
layout: "aws"
page_title: "AWS: aws_apprunner_hosted_zone_id"
description: |-
  Get AWS App Runner Hosted Zone Id
---

# Data Source: aws_apprunner_hosted_zone_id

Use this data source to get the HostedZoneId of an AWS App Runner service deployed
in a given region for the purpose of using it in an AWS Route53 Alias record.

## Example Usage

```terraform
data "aws_apprunner_hosted_zone_id" "main" {}

resource "aws_route53_record" "www" {
  zone_id = aws_route53_zone.primary.zone_id
  name    = "example.com"
  type    = "A"

  alias {
    name                   = aws_apprunner_custom_domain_association.main.dns_target
    zone_id                = data.aws_apprunner_hosted_zone_id.main.id
    evaluate_target_health = true
  }
}
```

## Argument Reference

* `region` - (Optional) Name of the region whose AWS App Runner service HostedZoneId is desired.
  Defaults to the region from the AWS provider configuration.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `id` - ID of the AWS App Runner service HostedZoneId in the selected region.
