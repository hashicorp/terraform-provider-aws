---
subcategory: "Network Firewall"
layout: "aws"
page_title: "AWS: aws_networkfirewall_proxy_rules_exclusive"
description: |-
  Manages AWS Network Firewall Proxy Rules Exclusive within a Proxy Rule Group.
---

# Resource: aws_networkfirewall_proxy_rules_exclusive

Manages AWS Network Firewall Proxy Rules Exclusive within a Proxy Rule Group. Proxy rules define conditions and actions for HTTP/HTTPS traffic inspection across three request/response phases: PRE_DNS, PRE_REQUEST, and POST_RESPONSE.

~> **NOTE:** This resource requires an existing [`aws_networkfirewall_proxy_rule_group`](networkfirewall_proxy_rule_group.html).

## Example Usage

### Basic Usage

```terraform
resource "aws_networkfirewall_proxy_rule_group" "example" {
  name = "example"
}

resource "aws_networkfirewall_proxy_rules_exclusive" "example" {
  proxy_rule_group_arn = aws_networkfirewall_proxy_rule_group.example.arn

  pre_dns {
    proxy_rule_name = "allow-example-com"
    action          = "ALLOW"

    conditions {
      condition_key      = "request:DestinationDomain"
      condition_operator = "StringEquals"
      condition_values   = ["example.com"]
    }
  }
}
```

### Multiple Rules Across Phases

```terraform
resource "aws_networkfirewall_proxy_rule_group" "example" {
  name = "example"
}

resource "aws_networkfirewall_proxy_rules_exclusive" "example" {
  proxy_rule_group_arn = aws_networkfirewall_proxy_rule_group.example.arn

  # DNS phase rules
  pre_dns {
    proxy_rule_name = "block-malicious-domains"
    action          = "DROP"
    description     = "Block known malicious domains"

    conditions {
      condition_key      = "request:DestinationDomain"
      condition_operator = "StringEquals"
      condition_values   = ["malicious.com", "badactor.net"]
    }
  }

  # Request phase rules
  pre_request {
    proxy_rule_name = "allow-api-requests"
    action          = "ALLOW"
    description     = "Allow API endpoint access"

    conditions {
      condition_key      = "request:Http:Uri"
      condition_operator = "StringEquals"
      condition_values   = ["/api/v1", "/api/v2"]
    }

    conditions {
      condition_key      = "request:Http:Method"
      condition_operator = "StringEquals"
      condition_values   = ["GET", "POST"]
    }
  }

  # Response phase rules
  post_response {
    proxy_rule_name = "block-large-responses"
    action          = "DROP"
    description     = "Block responses with status code >= 500"

    conditions {
      condition_key      = "response:Http:StatusCode"
      condition_operator = "NumericGreaterThanEquals"
      condition_values   = ["500"]
    }
  }
}
```

### Using Proxy Rule Group Name

```terraform
resource "aws_networkfirewall_proxy_rule_group" "example" {
  name = "example"
}

resource "aws_networkfirewall_proxy_rules_exclusive" "example" {
  proxy_rule_group_name = aws_networkfirewall_proxy_rule_group.example.name

  pre_dns {
    proxy_rule_name = "allow-corporate-domains"
    action          = "ALLOW"

    conditions {
      condition_key      = "request:DestinationDomain"
      condition_operator = "StringEquals"
      condition_values   = ["example.com", "example.org"]
    }
  }
}
```

## Argument Reference

The following arguments are optional:

* `post_response` - (Optional) Rules to apply during the POST_RESPONSE phase. See [Rule Configuration](#rule-configuration) below.
* `pre_dns` - (Optional) Rules to apply during the PRE_DNS phase. See [Rule Configuration](#rule-configuration) below.
* `pre_request` - (Optional) Rules to apply during the PRE_REQUEST phase. See [Rule Configuration](#rule-configuration) below.
* `proxy_rule_group_arn` - (Optional) ARN of the proxy rule group. Conflicts with `proxy_rule_group_name`. Required if `proxy_rule_group_name` is not specified.
* `proxy_rule_group_name` - (Optional) Name of the proxy rule group. Conflicts with `proxy_rule_group_arn`. Required if `proxy_rule_group_arn` is not specified.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

### Rule Configuration

Each rule block (`post_response`, `pre_dns`, `pre_request`) supports the following:

* `action` - (Required) Action to take when conditions match. Valid values: `ALLOW`, `DROP`.
* `conditions` - (Required) One or more condition blocks. See [Conditions](#conditions) below.
* `proxy_rule_name` - (Required) Unique name for the proxy rule within the rule group.
* `description` - (Optional) Description of the rule.
* `insert_position` - (Optional) Position to insert the rule. Rules are evaluated in order.

### Conditions

Each `conditions` block supports the following:

* `condition_key` - (Required) Attribute to evaluate. Valid values include:
    - Request-based: `request:SourceAccount`, `request:SourceVpc`, `request:SourceVpce`, `request:Time`, `request:SourceIp`, `request:DestinationIp`, `request:SourcePort`, `request:DestinationPort`, `request:Protocol`, `request:DestinationDomain`, `request:Http:Uri`, `request:Http:Method`, `request:Http:UserAgent`, `request:Http:ContentType`, `request:Http:Header/<CustomHeaderName>`
    - Response-based: `response:Http:StatusCode`, `response:Http:ContentType`, `response:Http:Header/<CustomHeaderName>`

  ~> **NOTE:** HTTP field matching for HTTPS requests requires TLS decryption to be enabled. Without TLS decryption, only IP-based filtering is available in the pre-request phase.
* `condition_operator` - (Required) Comparison operator. Valid values: `StringEquals`, `NumericGreaterThan`, `NumericGreaterThanEquals`.
* `condition_values` - (Required) List of values to compare against.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - ARN of the Proxy Rule Group.
* `proxy_rule_group_arn` - ARN of the Proxy Rule Group (computed if `proxy_rule_group_name` was provided).
* `proxy_rule_group_name` - Name of the Proxy Rule Group (computed if `proxy_rule_group_arn` was provided).

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Network Firewall Proxy Rules Exclusive using the `proxy_rule_group_arn`. For example:

```terraform
import {
  to = aws_networkfirewall_proxy_rules_exclusive.example
  id = "arn:aws:network-firewall:us-west-2:123456789012:proxy-rule-group/example"
}
```

Using `terraform import`, import Network Firewall Proxy Rules Exclusive using the `proxy_rule_group_arn`. For example:

```console
% terraform import aws_networkfirewall_proxy_rules_exclusive.example arn:aws:network-firewall:us-west-2:123456789012:proxy-rule-group/example
```
