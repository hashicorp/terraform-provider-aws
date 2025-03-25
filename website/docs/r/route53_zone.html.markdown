---
subcategory: "Route 53"
layout: "aws"
page_title: "AWS: aws_route53_zone"
description: |-
  Manages a Route53 Hosted Zone
---

# Resource: aws_route53_zone

Manages a Route53 Hosted Zone. For managing Domain Name System Security Extensions (DNSSEC), see the [`aws_route53_key_signing_key`](route53_key_signing_key.html) and [`aws_route53_hosted_zone_dnssec`](route53_hosted_zone_dnssec.html) resources.

## Example Usage

### Public Zone

```terraform
resource "aws_route53_zone" "primary" {
  name = "example.com"
}
```

### Public Subdomain Zone

For use in subdomains, note that you need to create a
`aws_route53_record` of type `NS` as well as the subdomain
zone.

```terraform
resource "aws_route53_zone" "main" {
  name = "example.com"
}

resource "aws_route53_zone" "dev" {
  name = "dev.example.com"

  tags = {
    Environment = "dev"
  }
}

resource "aws_route53_record" "dev-ns" {
  zone_id = aws_route53_zone.main.zone_id
  name    = "dev.example.com"
  type    = "NS"
  ttl     = "30"
  records = aws_route53_zone.dev.name_servers
}
```

### Private Zone

~> **NOTE:** Terraform provides both exclusive VPC associations defined in-line in this resource via `vpc` configuration blocks and a separate [Zone VPC Association](/docs/providers/aws/r/route53_zone_association.html) resource. At this time, you cannot use in-line VPC associations in conjunction with any `aws_route53_zone_association` resources with the same zone ID otherwise it will cause a perpetual difference in plan output. You can optionally use the generic Terraform resource [lifecycle configuration block](https://www.terraform.io/docs/configuration/meta-arguments/lifecycle.html) with `ignore_changes` to manage additional associations via the `aws_route53_zone_association` resource.

~> **NOTE:** Private zones require at least one VPC association at all times.

```terraform
resource "aws_route53_zone" "private" {
  name = "example.com"

  vpc {
    vpc_id = aws_vpc.example.id
  }
}
```

## Argument Reference

This resource supports the following arguments:

* `name` - (Required) This is the name of the hosted zone.
* `comment` - (Optional) A comment for the hosted zone. Defaults to 'Managed by Terraform'.
* `delegation_set_id` - (Optional) The ID of the reusable delegation set whose NS records you want to assign to the hosted zone. Conflicts with `vpc` as delegation sets can only be used for public zones.
* `force_destroy` - (Optional) Whether to destroy all records (possibly managed outside of Terraform) in the zone when destroying the zone.
* `tags` - (Optional) A map of tags to assign to the zone. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.
* `vpc` - (Optional) Configuration block(s) specifying VPC(s) to associate with a private hosted zone. Conflicts with the `delegation_set_id` argument in this resource and any [`aws_route53_zone_association` resource](/docs/providers/aws/r/route53_zone_association.html) specifying the same zone ID. Detailed below.

### vpc Argument Reference

* `vpc_id` - (Required) ID of the VPC to associate.
* `vpc_region` - (Optional) Region of the VPC to associate. Defaults to AWS provider region.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - The Amazon Resource Name (ARN) of the Hosted Zone.
* `zone_id` - The Hosted Zone ID. This can be referenced by zone records.
* `name_servers` - A list of name servers in associated (or default) delegation set.
  Find more about delegation sets in [AWS docs](https://docs.aws.amazon.com/Route53/latest/APIReference/actions-on-reusable-delegation-sets.html).
* `primary_name_server` - The Route 53 name server that created the SOA record.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `30m`)
* `update` - (Default `30m`)
* `delete` - (Default `30m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Route53 Zones using the zone `id`. For example:

```terraform
import {
  to = aws_route53_zone.myzone
  id = "Z1D633PJN98FT9"
}
```

Using `terraform import`, import Route53 Zones using the zone `id`. For example:

```console
% terraform import aws_route53_zone.myzone Z1D633PJN98FT9
```
