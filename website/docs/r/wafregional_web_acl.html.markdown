---
subcategory: "WAF Classic Regional"
layout: "aws"
page_title: "AWS: aws_wafregional_web_acl"
description: |-
  Provides a AWS WAF Regional web access control group (ACL) resource for use with ALB.
---

# Resource: aws_wafregional_web_acl

Provides a WAF Regional Web ACL Resource for use with Application Load Balancer.

## Example Usage

### Regular Rule

```terraform
resource "aws_wafregional_ipset" "ipset" {
  name = "tfIPSet"

  ip_set_descriptor {
    type  = "IPV4"
    value = "192.0.7.0/24"
  }
}

resource "aws_wafregional_rule" "wafrule" {
  name        = "tfWAFRule"
  metric_name = "tfWAFRule"

  predicate {
    data_id = aws_wafregional_ipset.ipset.id
    negated = false
    type    = "IPMatch"
  }
}

resource "aws_wafregional_web_acl" "wafacl" {
  name        = "tfWebACL"
  metric_name = "tfWebACL"

  default_action {
    type = "ALLOW"
  }

  rule {
    action {
      type = "BLOCK"
    }

    priority = 1
    rule_id  = aws_wafregional_rule.wafrule.id
    type     = "REGULAR"
  }
}
```

### Group Rule

```terraform
resource "aws_wafregional_web_acl" "example" {
  name        = "example"
  metric_name = "example"

  default_action {
    type = "ALLOW"
  }

  rule {
    priority = 1
    rule_id  = aws_wafregional_rule_group.example.id
    type     = "GROUP"

    override_action {
      type = "NONE"
    }
  }
}
```

### Logging

~> *NOTE:* The Kinesis Firehose Delivery Stream name must begin with `aws-waf-logs-`. See the [AWS WAF Developer Guide](https://docs.aws.amazon.com/waf/latest/developerguide/logging.html) for more information about enabling WAF logging.

```terraform
resource "aws_wafregional_web_acl" "example" {
  # ... other configuration ...

  logging_configuration {
    log_destination = aws_kinesis_firehose_delivery_stream.example.arn

    redacted_fields {
      field_to_match {
        type = "URI"
      }

      field_to_match {
        data = "referer"
        type = "HEADER"
      }
    }
  }
}
```

## Argument Reference

This resource supports the following arguments:

* `default_action` - (Required) Action that you want AWS WAF Regional to take when a request doesn't match the criteria in any of the rules that are associated with the web ACL. See [`default_action` Block](#default_action-block) below for details.
* `logging_configuration` - (Optional) Configuration block to enable WAF logging. See [`logging_configuration` Block](#logging_configuration-block) below for details.
* `metric_name` - (Required) Name or description for the Amazon CloudWatch metric of this web ACL.
* `name` - (Required) Name or description of the web ACL.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `rule` - (Optional) Set of configuration blocks containing rules for the web ACL. See [`rule` Block](#rule-block) below for details.
* `tags` - (Optional) Key-value map of resource tags. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

### `default_action` Block

The `default_action` block supports the following argument:

* `type` - (Required) How you want AWS WAF Regional to respond to requests that match the settings in a rule. For example, `ALLOW`, `BLOCK` or `COUNT`.

### `logging_configuration` Block

The `logging_configuration` block supports the following arguments:

* `log_destination` - (Required) ARN of Kinesis Firehose Delivery Stream.
* `redacted_fields` - (Optional) Configuration block containing parts of the request that you want redacted from the logs. See [`redacted_fields` Block](#redacted_fields-block) below for details.

### `redacted_fields` Block

The `redacted_fields` block supports the following argument:

* `field_to_match` - (Required) Set of configuration blocks for fields to redact. See [`field_to_match` Block](#field_to_match-block) below for details.

### `field_to_match` Block

-> Additional information about this configuration can be found in the [AWS WAF Regional API Reference](https://docs.aws.amazon.com/waf/latest/APIReference/API_regional_FieldToMatch.html).

The `field_to_match` block supports the following arguments:

* `data` - (Optional) When the value of `type` is `HEADER`, enter the name of the header that you want the WAF to search, for example, `User-Agent` or `Referer`. If the value of `type` is any other value, omit `data`.
* `type` - (Required) Part of the web request that you want AWS WAF to search for a specified string. For example, `HEADER` or `METHOD`.

### `rule` Block

-> Additional information about this configuration can be found in the [AWS WAF Regional API Reference](https://docs.aws.amazon.com/waf/latest/APIReference/API_regional_ActivatedRule.html).

The `rule` block supports the following arguments:

* `action` - (Optional) Configuration block of the action that CloudFront or AWS WAF takes when a web request matches the conditions in the rule. Not used if `type` is `GROUP`. See [`action` and `override_action` Block](#action-and-override_action-block) below for details.
* `override_action` - (Optional) Configuration block of the override the action that a group requests CloudFront or AWS WAF takes when a web request matches the conditions in the rule. Only used if `type` is `GROUP`. See [`action` and `override_action` Block](#action-and-override_action-block) below for details.
* `priority` - (Required) Order in which the rules in a WebACL are evaluated. Rules with a lower value are evaluated before rules with a higher value.
* `rule_id` - (Required) ID of the associated WAF (Regional) rule (e.g., [`aws_wafregional_rule`](/docs/providers/aws/r/wafregional_rule.html)). WAF (Global) rules cannot be used.
* `type` - (Optional) Rule type, either `REGULAR`, as defined by [Rule](http://docs.aws.amazon.com/waf/latest/APIReference/API_Rule.html), `RATE_BASED`, as defined by [RateBasedRule](http://docs.aws.amazon.com/waf/latest/APIReference/API_RateBasedRule.html), or `GROUP`, as defined by [RuleGroup](https://docs.aws.amazon.com/waf/latest/APIReference/API_RuleGroup.html). The default is REGULAR. If you add a RATE_BASED rule, you need to set `type` as `RATE_BASED`. If you add a GROUP rule, you need to set `type` as `GROUP`.

### `action` and `override_action` Block

The `action` and `override_action` blocks support the following argument:

* `type` - (Required) How you want AWS WAF Regional to respond to requests that match the settings in a rule. Valid values for `action` are `ALLOW`, `BLOCK` or `COUNT`. Valid values for `override_action` are `COUNT` and `NONE`.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the WAF Regional WebACL.
* `id` - ID of the WAF Regional WebACL.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import WAF Regional Web ACL using the ID. For example:

```terraform
import {
  to = aws_wafregional_web_acl.wafacl
  id = "a1b2c3d4-d5f6-7777-8888-9999aaaabbbbcccc"
}
```

Using `terraform import`, import WAF Regional Web ACL using the ID. For example:

```console
% terraform import aws_wafregional_web_acl.wafacl a1b2c3d4-d5f6-7777-8888-9999aaaabbbbcccc
```
