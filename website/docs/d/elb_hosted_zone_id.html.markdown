---
subcategory: "ELB Classic"
layout: "aws"
page_title: "AWS: aws_elb_hosted_zone_id"
description: |-
  Get AWS Elastic Load Balancing Hosted Zone Id
---

# Data Source: aws_elb_hosted_zone_id

Use this data source to get the HostedZoneId of the AWS Elastic Load Balancing HostedZoneId
in a given region for the purpose of using in an AWS Route53 Alias.

## Example Usage

```terraform
data "aws_elb_hosted_zone_id" "main" {}

resource "aws_route53_record" "www" {
  zone_id = aws_route53_zone.primary.zone_id
  name    = "example.com"
  type    = "A"

  alias {
    name                   = aws_elb.main.dns_name
    zone_id                = data.aws_elb_hosted_zone_id.main.id
    evaluate_target_health = true
  }
}
```

## Argument Reference

This data source supports the following arguments:

* `region` - (Optional) Name of the Region whose AWS ELB HostedZoneId is desired. Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `id` - ID of the AWS ELB HostedZoneId in the selected Region.
