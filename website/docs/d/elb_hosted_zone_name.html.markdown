---
layout: "aws"
page_title: "AWS: aws_elb_hosted_zone_name"
sidebar_current: "docs-aws-datasource-elb-hosted-zone-name"
description: |-
  Get AWS Elastic Load Balancing Hosted Zone Name and ID
---

# aws\_elb\_hosted\_zone\_name

Use this data source to get the HostedZoneName and HostedZoneId of an existing AWS Elastic Load Balancer.

## Example Usage

```hcl
data "aws_elb_hosted_zone_name" "main" {
  name = "my-elb"
}

resource "aws_route53_record" "www" {
  zone_id = "${aws_route53_zone.primary.zone_id}"
  name    = "example.com"
  type    = "A"

  alias {
    name                   = "${data.aws_elb_hosted_zone_id.main.hosted_zone_name}"
    zone_id                = "${data.aws_elb_hosted_zone_id.main.hosted_zone_id}"
    evaluate_target_health = true
  }
}
```

## Argument Reference

* `name` - (Required) Name of the ELB.

## Attributes Reference

* `hosted_zone_id` - The ID of the AWS ELB's HostedZoneId.
* `hosted_zone_name` - The Hosted Zone Name of the AWS ELB.
* `dns_name` - The DNS Name of the AWS ELB.
