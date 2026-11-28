---
subcategory: "WAF Classic Regional"
layout: "aws"
page_title: "AWS: aws_wafregional_sql_injection_match_set"
description: |-
  Provides a AWS WAF Regional SqlInjectionMatchSet resource for use with ALB.
---

# Resource: aws_wafregional_sql_injection_match_set

Provides a WAF Regional SQL Injection Match Set Resource for use with Application Load Balancer.

## Example Usage

```terraform
resource "aws_wafregional_sql_injection_match_set" "sql_injection_match_set" {
  name = "tf-sql_injection_match_set"

  sql_injection_match_tuple {
    text_transformation = "URL_DECODE"

    field_to_match {
      type = "QUERY_STRING"
    }
  }
}
```

## Argument Reference

This resource supports the following arguments:

* `account` - (Optional) AWS account ID where this resource will be managed. Defaults to the account of the provider's credentials. See the [Per-Resource Account guide](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/guides/per-resource-account).
* `name` - (Required) Name or description of the SQL Injection Match Set.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `sql_injection_match_tuple` - (Optional) Parts of web requests that you want AWS WAF to inspect for malicious SQL code and, if you want AWS WAF to inspect a header, the name of the header. See [`sql_injection_match_tuple` Block](#sql_injection_match_tuple-block) below for details.

### `sql_injection_match_tuple` Block

The `sql_injection_match_tuple` block supports the following arguments:

* `field_to_match` - (Required) Where in a web request to look for snippets of malicious SQL code. See [`field_to_match` Block](#field_to_match-block) below for details.
* `text_transformation` - (Required) Text transformations used to eliminate unusual formatting that attackers use in web requests in an effort to bypass AWS WAF. If you specify a transformation, AWS WAF performs the transformation on `field_to_match` before inspecting a request for a match. For example, `CMD_LINE`, `HTML_ENTITY_DECODE` or `NONE`. See [docs](https://docs.aws.amazon.com/waf/latest/APIReference/API_regional_SqlInjectionMatchTuple.html#WAF-Type-regional_SqlInjectionMatchTuple-TextTransformation) for all supported values.

### `field_to_match` Block

The `field_to_match` block supports the following arguments:

* `data` - (Optional) When `type` is `HEADER`, enter the name of the header that you want to search, e.g., `User-Agent` or `Referer`. If `type` is any other value, omit this field.
* `type` - (Required) Part of the web request that you want AWS WAF to search for a specified string. For example, `HEADER`, `METHOD` or `BODY`. See [docs](https://docs.aws.amazon.com/waf/latest/APIReference/API_regional_FieldToMatch.html) for all supported values.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - ID of the WAF SqlInjectionMatchSet.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import WAF Regional SQL Injection Match Set using the ID. For example:

```terraform
import {
  to = aws_wafregional_sql_injection_match_set.sql_injection_match_set
  id = "a1b2c3d4-d5f6-7777-8888-9999aaaabbbbcccc"
}
```

Using `terraform import`, import WAF Regional SQL Injection Match Set using the ID. For example:

```console
% terraform import aws_wafregional_sql_injection_match_set.sql_injection_match_set a1b2c3d4-d5f6-7777-8888-9999aaaabbbbcccc
```
