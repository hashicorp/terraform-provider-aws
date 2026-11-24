---
subcategory: "Route 53"
layout: "aws"
page_title: "AWS: aws_route53_record_set_protection"
description: |-
  Manages a Route 53 record set protection, which blocks changes to every record with a given name and type in a hosted zone.
---

# Resource: aws_route53_record_set_protection

Manages a Route 53 record set protection, which blocks changes to every record with a given name and type in a hosted zone. While the protection exists, Route 53 rejects any change that creates, updates, or deletes those records, for every principal, including those with `route53:*`. For more information, see [Protecting records from changes](https://docs.aws.amazon.com/Route53/latest/DeveloperGuide/rrset-protection.html) in the Amazon Route 53 Developer Guide.

~> A protection covers every record with its name and type, whatever the set identifier. To change a protected record, remove the protection, change the record, and then add the protection back, in separate applies. Changing a protected record in the same apply fails with an `InvalidChangeBatch` error.

~> A hosted zone that has protections can't be deleted, and [`aws_route53_zone`](route53_zone.html) `force_destroy` doesn't remove them. Reference the zone and the protected records from this resource, as in the examples, so that Terraform removes the protection first on destroy.

## Example Usage

### Protect Existing Records

```terraform
resource "aws_route53_zone" "example" {
  name = "example.com"
}

resource "aws_route53_record" "apex_mx" {
  zone_id = aws_route53_zone.example.zone_id
  name    = aws_route53_zone.example.name
  type    = "MX"
  ttl     = 300
  records = ["10 mail.example.com"]
}

resource "aws_route53_record_set_protection" "apex_mx" {
  zone_id     = aws_route53_zone.example.zone_id
  record_name = aws_route53_record.apex_mx.name
  record_type = aws_route53_record.apex_mx.type
  comment     = "Production mail. Contact the DNS team before changing."
}
```

### Reserve a Name and Type

A protection can exist for a name and type that has no records. No one can create records with that name and type until the protection is removed.

```terraform
resource "aws_route53_record_set_protection" "reserved" {
  zone_id     = aws_route53_zone.example.zone_id
  record_name = "legacy.example.com"
  record_type = "A"
  comment     = "Reserved. Do not reuse."
}
```

## Argument Reference

The following arguments are required:

* `record_name` - (Required) Fully qualified name of the records to protect, such as `www.example.com`. The name must be the hosted zone's name or a subdomain of it. A trailing dot, letter case, and `*` written as `\052` don't cause differences. Changing this forces a new resource to be created.
* `record_type` - (Required) DNS type of the records to protect, such as `A` or `MX`. Valid values are the types that [`aws_route53_record`](route53_record.html) accepts. Changing this forces a new resource to be created.
* `zone_id` - (Required) ID of the hosted zone. Changing this forces a new resource to be created.

The following arguments are optional:

* `comment` - (Optional) Reason for the protection, up to 256 characters. Route 53 includes it in the error message when it rejects a change. Changing this forces a new resource to be created.

## Attribute Reference

This resource exports no additional attributes.

## Import

In Terraform v1.12.0 and later, the [`import` block](https://developer.hashicorp.com/terraform/language/import) can be used with the `identity` attribute. For example:

```terraform
import {
  to = aws_route53_record_set_protection.example
  identity = {
    zone_id     = "Z1D633PJN98FT9"
    record_name = "www.example.com"
    record_type = "A"
  }
}

resource "aws_route53_record_set_protection" "example" {
  ### Configuration omitted for brevity ###
}
```

### Identity Schema

#### Required

* `record_name` (String) Name of the protected records.
* `record_type` (String) DNS type of the protected records.
* `zone_id` (String) ID of the hosted zone.

#### Optional

* `account_id` (String) AWS Account where this resource is managed.

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import a record set protection using the zone ID, record name, and record type separated by commas (`,`). For example:

```terraform
import {
  to = aws_route53_record_set_protection.example
  id = "Z1D633PJN98FT9,www.example.com,A"
}
```

Using `terraform import`, import a record set protection using the zone ID, record name, and record type separated by commas (`,`). For example:

```console
% terraform import aws_route53_record_set_protection.example Z1D633PJN98FT9,www.example.com,A
```
