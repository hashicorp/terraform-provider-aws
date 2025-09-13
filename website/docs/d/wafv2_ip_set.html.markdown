---
subcategory: "WAF"
layout: "aws"
page_title: "AWS: aws_wafv2_ip_set"
description: |-
  Retrieves the summary of a WAFv2 IP Set.
---

# Data Source: aws_wafv2_ip_set

Retrieves the summary of a WAFv2 IP Set.

## Example Usage

```terraform
data "aws_wafv2_ip_set" "example" {
  name  = "some-ip-set"
  scope = "REGIONAL"
}
```

## Argument Reference

This data source supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `name` - (Required) Name of the WAFv2 IP Set.
* `scope` - (Required) Specifies whether this is for an AWS CloudFront distribution or for a regional application. Valid values are `CLOUDFRONT` or `REGIONAL`. To work with CloudFront, you must also specify the region `us-east-1` (N. Virginia) on the AWS provider.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `addresses` - An array of strings that specifies zero or more IP addresses or blocks of IP addresses in Classless Inter-Domain Routing (CIDR) notation.
* `arn` - ARN of the entity.
* `description` - Description of the set that helps with identification.
* `id` - Unique identifier for the set.
* `ip_address_version` - IP address version of the set.
