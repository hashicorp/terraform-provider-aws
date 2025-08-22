---
subcategory: "ELB (Elastic Load Balancing)"
layout: "aws"
page_title: "AWS: aws_lb_hosted_zone_id"
description: |-
  Provides AWS Elastic Load Balancing Hosted Zone Id
---

# Data Source: aws_lb_hosted_zone_id

Use this data source to get the HostedZoneId of the AWS Elastic Load Balancing (ELB) in a given region for the purpose of using in an AWS Route53 Alias. Specify the ELB type (`network` or `application`) to return the relevant the associated HostedZoneId. Ref: [ELB service endpoints](https://docs.aws.amazon.com/general/latest/gr/elb.html#elb_region)

## Example Usage

```terraform
data "aws_lb_hosted_zone_id" "main" {}

resource "aws_route53_record" "www" {
  zone_id = aws_route53_zone.primary.zone_id
  name    = "example.com"
  type    = "A"

  alias {
    name                   = aws_lb.main.dns_name
    zone_id                = data.aws_lb_hosted_zone_id.main.id
    evaluate_target_health = true
  }
}
```

## Argument Reference

This data source supports the following arguments:

* `region` - (Optional) Name of the Region whose AWS ELB HostedZoneId is desired. Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `load_balancer_type` - (Optional) Type of load balancer to create. Possible values are `application` or `network`. The default value is `application`.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `id` - ID of the AWS ELB HostedZoneId in the selected Region.
