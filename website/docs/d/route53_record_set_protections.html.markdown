---
subcategory: "Route 53"
layout: "aws"
page_title: "AWS: aws_route53_record_set_protections"
description: |-
  Provides details about the record set protections in a Route 53 hosted zone.
---

# Data Source: aws_route53_record_set_protections

Provides details about the record set protections in a Route 53 hosted zone. While a protection exists, Route 53 rejects any change that creates, updates, or deletes the records with that name and type. To manage a protection, use the [`aws_route53_record_set_protection`](/docs/providers/aws/r/route53_record_set_protection.html) resource.

## Example Usage

```terraform
data "aws_route53_zone" "example" {
  name = "example.com"
}

data "aws_route53_record_set_protections" "example" {
  zone_id = data.aws_route53_zone.example.zone_id
}
```

## Argument Reference

This data source supports the following arguments:

* `zone_id` - (Required) ID of the hosted zone.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `protections` - List of protections in the hosted zone, sorted by record name and then by record type. Empty if the hosted zone has none. See [`protections`](#protections) below.

### `protections`

* `comment` - Comment on the protection.
* `record_name` - Fully qualified name of the protected record set.
* `record_type` - DNS type of the protected record set.
