---
subcategory: "WAF Classic Regional"
layout: "aws"
page_title: "AWS: aws_wafregional_xss_match_set"
description: |-
  Provides an AWS WAF Regional XSS Match Set resource for use with ALB.
---

# Resource: aws_wafregional_xss_match_set

Provides a WAF Regional XSS Match Set Resource for use with Application Load Balancer.

## Example Usage

```terraform
resource "aws_wafregional_xss_match_set" "xss_match_set" {
  name = "xss_match_set"

  xss_match_tuple {
    text_transformation = "NONE"

    field_to_match {
      type = "URI"
    }
  }

  xss_match_tuple {
    text_transformation = "NONE"

    field_to_match {
      type = "QUERY_STRING"
    }
  }
}
```

## Argument Reference

This resource supports the following arguments:

* `account` - (Optional) AWS account ID where this resource will be managed. Defaults to the account of the provider's credentials. See the [Per-Resource Account guide](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/guides/per-resource-account).
* `name` - (Required) Name of the set.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `xss_match_tuple` - (Optional) Parts of web requests that you want to inspect for cross-site scripting attacks. See [`xss_match_tuple` Block](#xss_match_tuple-block) below for details.

### `xss_match_tuple` Block

The `xss_match_tuple` block supports the following arguments:

* `field_to_match` - (Required) Where in a web request to look for cross-site scripting attacks. See [`field_to_match` Block](#field_to_match-block) below for details.
* `text_transformation` - (Required) Which text transformation, if any, to perform on the web request before inspecting the request for cross-site scripting attacks.

### `field_to_match` Block

The `field_to_match` block supports the following arguments:

* `data` - (Optional) When the value of `type` is `HEADER`, enter the name of the header that you want the WAF to search, for example, `User-Agent` or `Referer`. If the value of `type` is any other value, omit `data`.
* `type` - (Required) Part of the web request that you want AWS WAF to search for a specified string. For example, `HEADER` or `METHOD`.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - ID of the Regional WAF XSS Match Set.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import AWS WAF Regional XSS Match using the `id`. For example:

```terraform
import {
  to = aws_wafregional_xss_match_set.example
  id = "12345abcde"
}
```

Using `terraform import`, import AWS WAF Regional XSS Match using the `id`. For example:

```console
% terraform import aws_wafregional_xss_match_set.example 12345abcde
```
