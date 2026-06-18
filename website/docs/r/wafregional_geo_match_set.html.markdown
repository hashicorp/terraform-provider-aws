---
subcategory: "WAF Classic Regional"
layout: "aws"
page_title: "AWS: aws_wafregional_geo_match_set"
description: |-
  Provides a AWS WAF Regional Geo Match Set resource.
---

# Resource: aws_wafregional_geo_match_set

Provides a WAF Regional Geo Match Set Resource

## Example Usage

```terraform
resource "aws_wafregional_geo_match_set" "geo_match_set" {
  name = "geo_match_set"

  geo_match_constraint {
    type  = "Country"
    value = "US"
  }

  geo_match_constraint {
    type  = "Country"
    value = "CA"
  }
}
```

## Argument Reference

This resource supports the following arguments:

* `geo_match_constraint` - (Optional) Geo Match Constraint objects which contain the country that you want AWS WAF to search for.
* `name` - (Required) Name or description of the Geo Match Set.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

### `geo_match_constraint` Block

* `type` - (Required) Type of geographical area you want AWS WAF to search for. Currently `Country` is the only valid value.
* `value` - (Required) Two-letter country code that you want AWS WAF to search for, e.g., `US`, `CA`, `RU`, `CN`. See [docs](https://docs.aws.amazon.com/waf/latest/APIReference/API_GeoMatchConstraint.html) for all supported values.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - ID of the WAF Regional Geo Match Set.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import WAF Regional Geo Match Set using the id. For example:

```terraform
import {
  to = aws_wafregional_geo_match_set.geo_match_set
  id = "a1b2c3d4-d5f6-7777-8888-9999aaaabbbbcccc"
}
```

Using `terraform import`, import WAF Regional Geo Match Set using the id. For example:

```console
% terraform import aws_wafregional_geo_match_set.geo_match_set a1b2c3d4-d5f6-7777-8888-9999aaaabbbbcccc
```
