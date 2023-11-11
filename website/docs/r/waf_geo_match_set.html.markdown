---
subcategory: "WAF Classic"
layout: "aws"
page_title: "AWS: aws_waf_geo_match_set"
description: |-
  Provides a AWS WAF GeoMatchSet resource.
---

# Resource: aws_waf_geo_match_set

Provides a WAF Geo Match Set Resource

## Example Usage

```terraform
resource "aws_waf_geo_match_set" "geo_match_set" {
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

* `name` - (Required) The name or description of the GeoMatchSet.
* `geo_match_constraint` - (Optional) The GeoMatchConstraint objects which contain the country that you want AWS WAF to search for.

## Nested Blocks

### `geo_match_constraint`

#### Arguments

* `type` - (Required) The type of geographical area you want AWS WAF to search for. Currently Country is the only valid value.
* `value` - (Required) The country that you want AWS WAF to search for.
  This is the two-letter country code, e.g., `US`, `CA`, `RU`, `CN`, etc.
  See [docs](https://docs.aws.amazon.com/waf/latest/APIReference/API_GeoMatchConstraint.html) for all supported values.

## Remarks

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - The ID of the WAF GeoMatchSet.
* `arn` - Amazon Resource Name (ARN)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import WAF Geo Match Set using their ID. For example:

```terraform
import {
  to = aws_waf_geo_match_set.example
  id = "a1b2c3d4-d5f6-7777-8888-9999aaaabbbbcccc"
}
```

Using `terraform import`, import WAF Geo Match Set using their ID. For example:

```console
% terraform import aws_waf_geo_match_set.example a1b2c3d4-d5f6-7777-8888-9999aaaabbbbcccc
```
