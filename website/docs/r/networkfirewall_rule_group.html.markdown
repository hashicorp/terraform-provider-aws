---
subcategory: "Network Firewall"
layout: "aws"
page_title: "AWS: aws_networkfirewall_rule_group"
description: |-
  Provides an AWS Network Firewall Rule Group resource.
---

# Resource: aws_networkfirewall_rule_group

Provides an AWS Network Firewall Rule Group Resource

## Example Usage

### Stateful Inspection for denying access to a domain

```terraform
resource "aws_networkfirewall_rule_group" "example" {
  capacity = 100
  name     = "example"
  type     = "STATEFUL"
  rule_group {
    rules_source {
      rules_source_list {
        generated_rules_type = "DENYLIST"
        target_types         = ["HTTP_HOST"]
        targets              = ["test.example.com"]
      }
    }
  }

  tags = {
    Tag1 = "Value1"
    Tag2 = "Value2"
  }
}
```

### Stateful Inspection for permitting packets from a source IP address

```terraform
resource "aws_networkfirewall_rule_group" "example" {
  capacity    = 50
  description = "Permits http traffic from source"
  name        = "example"
  type        = "STATEFUL"
  rule_group {
    rules_source {
      dynamic "stateful_rule" {
        for_each = local.ips
        content {
          action = "PASS"
          header {
            destination      = "ANY"
            destination_port = "ANY"
            protocol         = "HTTP"
            direction        = "ANY"
            source_port      = "ANY"
            source           = stateful_rule.value
          }
          rule_option {
            keyword  = "sid"
            settings = ["1"]
          }
        }
      }
    }
  }

  tags = {
    Name = "permit HTTP from source"
  }
}

locals {
  ips = ["1.1.1.1/32", "1.0.0.1/32"]
}
```

### Stateful Inspection for blocking packets from going to an intended destination

```terraform
resource "aws_networkfirewall_rule_group" "example" {
  capacity = 100
  name     = "example"
  type     = "STATEFUL"
  rule_group {
    rules_source {
      stateful_rule {
        action = "DROP"
        header {
          destination      = "124.1.1.24/32"
          destination_port = 53
          direction        = "ANY"
          protocol         = "TCP"
          source           = "1.2.3.4/32"
          source_port      = 53
        }
        rule_option {
          keyword  = "sid"
          settings = ["1"]
        }
      }
    }
  }

  tags = {
    Tag1 = "Value1"
    Tag2 = "Value2"
  }
}
```

### Stateful Inspection from rules specifications defined in Suricata flat format

```terraform
resource "aws_networkfirewall_rule_group" "example" {
  capacity = 100
  name     = "example"
  type     = "STATEFUL"
  rules    = file("example.rules")

  tags = {
    Tag1 = "Value1"
    Tag2 = "Value2"
  }
}
```

### Stateful Inspection from rule group specifications using rule variables and Suricata format rules

```terraform
resource "aws_networkfirewall_rule_group" "example" {
  capacity = 100
  name     = "example"
  type     = "STATEFUL"
  rule_group {
    rule_variables {
      ip_sets {
        key = "WEBSERVERS_HOSTS"
        ip_set {
          definition = ["10.0.0.0/16", "10.0.1.0/24", "192.168.0.0/16"]
        }
      }
      ip_sets {
        key = "EXTERNAL_HOST"
        ip_set {
          definition = ["1.2.3.4/32"]
        }
      }
      port_sets {
        key = "HTTP_PORTS"
        port_set {
          definition = ["443", "80"]
        }
      }
    }
    rules_source {
      rules_string = file("suricata_rules_file")
    }
  }
  tags = {
    Tag1 = "Value1"
    Tag2 = "Value2"
  }
}
```

### Stateless Inspection with a Custom Action

```terraform
resource "aws_networkfirewall_rule_group" "example" {
  description = "Stateless Rate Limiting Rule"
  capacity    = 100
  name        = "example"
  type        = "STATELESS"
  rule_group {
    rules_source {
      stateless_rules_and_custom_actions {
        custom_action {
          action_definition {
            publish_metric_action {
              dimension {
                value = "2"
              }
            }
          }
          action_name = "ExampleMetricsAction"
        }
        stateless_rule {
          priority = 1
          rule_definition {
            actions = ["aws:pass", "ExampleMetricsAction"]
            match_attributes {
              source {
                address_definition = "1.2.3.4/32"
              }
              source_port {
                from_port = 443
                to_port   = 443
              }
              destination {
                address_definition = "124.1.1.5/32"
              }
              destination_port {
                from_port = 443
                to_port   = 443
              }
              protocols = [6]
              tcp_flag {
                flags = ["SYN"]
                masks = ["SYN", "ACK"]
              }
            }
          }
        }
      }
    }
  }

  tags = {
    Tag1 = "Value1"
    Tag2 = "Value2"
  }
}
```

### IP Set References to the Rule Group

```terraform
resource "aws_networkfirewall_rule_group" "example" {
  capacity = 100
  name     = "example"
  type     = "STATEFUL"
  rule_group {
    rules_source {
      rules_source_list {
        generated_rules_type = "DENYLIST"
        target_types         = ["HTTP_HOST"]
        targets              = ["test.example.com"]
      }
    }
    reference_sets {
      ip_set_references {
        key = "example"
        ip_set_reference {
          reference_arn = aws_ec2_managed_prefix_list.this.arn
        }
      }
    }
  }

  tags = {
    Tag1 = "Value1"
    Tag2 = "Value2"
  }
}
```

### Example with S3 as source for the suricata rules

```terraform

data "aws_s3_object" "suricata_rules" {
  bucket = aws_s3_bucket.suricata_rules.id
  key    = "rules/custom.rules"
}

resource "aws_networkfirewall_rule_group" "s3_rules_example" {
  capacity = 1000
  name     = "my-terraform-s3-rules"
  type     = "STATEFUL"

  rule_group {
    rule_variables {
      ip_sets {
        key = "HOME_NET"
        ip_set {
          definition = ["10.0.0.0/16", "192.168.0.0/16", "172.16.0.0/12"]
        }
      }
      port_sets {
        key = "HTTP_PORTS"
        port_set {
          definition = ["443", "80"]
        }
      }
    }

    rules_source {
      rules_string = data.aws_s3_object.suricata_rules.body
    }
  }

  tags = {
    ManagedBy = "terraform"
  }

}

```

## Argument Reference

This resource supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `capacity` - (Required, Forces new resource) The maximum number of operating resources that this rule group can use. For a stateless rule group, the capacity required is the sum of the capacity requirements of the individual rules. For a stateful rule group, the minimum capacity required is the number of individual rules.
* `description` - (Optional) A friendly description of the rule group.
* `encryption_configuration` - (Optional) KMS encryption configuration settings. See [Encryption Configuration](#encryption-configuration) below for details.
* `name` - (Required, Forces new resource) A friendly name of the rule group.
* `rule_group` - (Optional) A configuration block that defines the rule group rules. Required unless `rules` is specified. See [Rule Group](#rule-group) below for details.
* `rules` - (Optional) The stateful rule group rules specifications in Suricata file format, with one rule per line. Use this to import your existing Suricata compatible rule groups. Required unless `rule_group` is specified.
* `tags` - (Optional) A map of key:value pairs to associate with the resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.
* `type` - (Required) Whether the rule group is stateless (containing stateless rules) or stateful (containing stateful rules). Valid values include: `STATEFUL` or `STATELESS`.

### Encryption Configuration

`encryption_configuration` settings for customer managed KMS keys. Remove this block to use the default AWS-managed KMS encryption (rather than setting `type` to `AWS_OWNED_KMS_KEY`).

* `key_id` - (Optional) The ID of the customer managed key. You can use any of the [key identifiers](https://docs.aws.amazon.com/kms/latest/developerguide/concepts.html#key-id) that KMS supports, unless you're using a key that's managed by another account. If you're using a key managed by another account, then specify the key ARN.
* `type` - (Required) The type of AWS KMS key to use for encryption of your Network Firewall resources. Valid values are `CUSTOMER_KMS` and `AWS_OWNED_KMS_KEY`.

### Rule Group

The `rule_group` block supports the following argument:

* `reference_sets` - (Optional) A configuration block that defines the IP Set References for the rule group. See [Reference Sets](#reference-sets) below for details. Please notes that there can only be a maximum of 5 `reference_sets` in a `rule_group`. See the [AWS documentation](https://docs.aws.amazon.com/network-firewall/latest/developerguide/rule-groups-ip-set-references.html#rule-groups-ip-set-reference-limits) for details.

* `rule_variables` - (Optional) A configuration block that defines additional settings available to use in the rules defined in the rule group. Can only be specified for **stateful** rule groups. See [Rule Variables](#rule-variables) below for details.

* `rules_source` - (Required) A configuration block that defines the stateful or stateless rules for the rule group. See [Rules Source](#rules-source) below for details.

* `stateful_rule_options` - (Optional) A configuration block that defines stateful rule options for the rule group. See [Stateful Rule Options](#stateful-rule-options) below for details.

### Reference Sets

The `reference_sets` block supports the following arguments:

* `ip_set_reference` - (Optional) Set of configuration blocks that define the IP Reference information. See [IP Set Reference](#ip-set-reference) below for details.

### Rule Variables

The `rule_variables` block supports the following arguments:

* `ip_sets` - (Optional) Set of configuration blocks that define IP address information. See [IP Sets](#ip-sets) below for details.

* `port_sets` - (Optional) Set of configuration blocks that define port range information. See [Port Sets](#port-sets) below for details.

### IP Sets

The `ip_sets` block supports the following arguments:

* `key` - (Required) A unique alphanumeric string to identify the `ip_set`.

* `ip_set` - (Required) A configuration block that defines a set of IP addresses. See [IP Set](#ip-set) below for details.

### IP Set

The `ip_set` configuration block supports the following argument:

* `definition` - (Required) Set of IP addresses and address ranges, in CIDR notation.

### IP Set Reference

The `ip_set_reference` configuration block supports the following argument:

* `key` - (Required) A unique alphanumeric string to identify the `ip_set`.

* `reference_arn` - (Required) Set of Managed Prefix IP ARN(s)

### Port Sets

The `port_sets` block supports the following arguments:

* `key` - (Required) An unique alphanumeric string to identify the `port_set`.

* `port_set` - (Required) A configuration block that defines a set of port ranges. See [Port Set](#port-set) below for details.

### Port Set

The `port_set` configuration block suppports the following argument:

* `definition` - (Required) Set of port ranges.

### Rules Source

The `rules_source` block supports the following arguments:

~> **NOTE:** Only one of `rules_source_list`, `rules_string`, `stateful_rule`, or `stateless_rules_and_custom_actions` must be specified.

* `rules_source_list` - (Optional) A configuration block containing **stateful** inspection criteria for a domain list rule group. See [Rules Source List](#rules-source-list) below for details.

* `rules_string` - (Optional) Stateful inspection criteria, provided in Suricata compatible rules. These rules contain the inspection criteria and the action to take for traffic that matches the criteria, so this type of rule group doesn’t have a separate action setting.

* `stateful_rule` - (Optional) Set of configuration blocks containing **stateful** inspection criteria for 5-tuple rules to be used together in a rule group. See [Stateful Rule](#stateful-rule) below for details.

* `stateless_rules_and_custom_actions` - (Optional) A configuration block containing **stateless** inspection criteria for a stateless rule group. See [Stateless Rules and Custom Actions](#stateless-rules-and-custom-actions) below for details.

### Stateful Rule Options

The `stateful_rule_options` block supports the following argument:

~> **NOTE:** If the `STRICT_ORDER` rule order is specified, this rule group can only be referenced in firewall policies that also utilize `STRICT_ORDER` for the stateful engine. `STRICT_ORDER` can only be specified when using a `rules_source` of `rules_string` or `stateful_rule`.

* `rule_order` - (Required) Indicates how to manage the order of the rule evaluation for the rule group. Default value: `DEFAULT_ACTION_ORDER`. Valid values: `DEFAULT_ACTION_ORDER`, `STRICT_ORDER`.

### Rules Source List

The `rules_source_list` block supports the following arguments:

* `generated_rules_type` - (Required) String value to specify whether domains in the target list are allowed or denied access. Valid values: `ALLOWLIST`, `DENYLIST`.

* `target_types` - (Required) Set of types of domain specifications that are provided in the `targets` argument. Valid values: `HTTP_HOST`, `TLS_SNI`.

* `targets` - (Required) Set of domains that you want to inspect for in your traffic flows.

### Stateful Rule

The `stateful_rule` block supports the following arguments:

* `action` - (Required) Action to take with packets in a traffic flow when the flow matches the stateful rule criteria. For all actions, AWS Network Firewall performs the specified action and discontinues stateful inspection of the traffic flow. Valid values: `ALERT`, `DROP`, `PASS`, or `REJECT`.

* `header` - (Required) A configuration block containing the stateful 5-tuple inspection criteria for the rule, used to inspect traffic flows. See [Header](#header) below for details.

* `rule_option` - (Required) Set of configuration blocks containing additional settings for a stateful rule. See [Rule Option](#rule-option) below for details.

### Stateless Rules and Custom Actions

The `stateless_rules_and_custom_actions` block supports the following arguments:

* `custom_action` - (Optional) Set of configuration blocks containing custom action definitions that are available for use by the set of `stateless rule`. See [Custom Action](#custom-action) below for details.

* `stateless_rule` - (Required) Set of configuration blocks containing the stateless rules for use in the stateless rule group. See [Stateless Rule](#stateless-rule) below for details.

### Header

The `header` block supports the following arguments:

* `destination` - (Required) The destination IP address or address range to inspect for, in CIDR notation. To match with any address, specify `ANY`.

* `destination_port` - (Required) The destination port to inspect for. To match with any address, specify `ANY`.

* `direction` - (Required) The direction of traffic flow to inspect. Valid values: `ANY` or `FORWARD`.

* `protocol` - (Required) The protocol to inspect. Valid values: `IP`, `TCP`, `UDP`, `ICMP`, `HTTP`, `FTP`, `TLS`, `SMB`, `DNS`, `DCERPC`, `SSH`, `SMTP`, `IMAP`, `MSN`, `KRB5`, `IKEV2`, `TFTP`, `NTP`, `DHCP`.

* `source` - (Required) The source IP address or address range for, in CIDR notation. To match with any address, specify `ANY`.

* `source_port` - (Required) The source port to inspect for. To match with any address, specify `ANY`.

### Rule Option

The `rule_option` block supports the following arguments:

* `keyword` - (Required) Keyword defined by open source detection systems like Snort or Suricata for stateful rule inspection.
See [Snort General Rule Options](http://manual-snort-org.s3-website-us-east-1.amazonaws.com/node31.html) or [Suricata Rule Options](https://suricata.readthedocs.io/en/suricata-5.0.1/rules/intro.html#rule-options) for more details.
* `settings` - (Optional) Set of strings for additional settings to use in stateful rule inspection.

### Custom Action

The `custom_action` block supports the following arguments:

* `action_definition` - (Required) A configuration block describing the custom action associated with the `action_name`. See [Action Definition](#action-definition) below for details.

* `action_name` - (Required, Forces new resource) A friendly name of the custom action.

### Stateless Rule

The `stateless_rule` block supports the following arguments:

* `priority` - (Required) A setting that indicates the order in which to run this rule relative to all of the rules that are defined for a stateless rule group. AWS Network Firewall evaluates the rules in a rule group starting with the lowest priority setting.

* `rule_definition` - (Required) A configuration block defining the stateless 5-tuple packet inspection criteria and the action to take on a packet that matches the criteria. See [Rule Definition](#rule-definition) below for details.

### Rule Definition

The `rule_definition` block supports the following arguments:

* `actions` - (Required) Set of actions to take on a packet that matches one of the stateless rule definition's `match_attributes`. For every rule you must specify 1 standard action, and you can add custom actions. Standard actions include: `aws:pass`, `aws:drop`, `aws:forward_to_sfe`.

* `match_attributes` - (Required) A configuration block containing criteria for AWS Network Firewall to use to inspect an individual packet in stateless rule inspection. See [Match Attributes](#match-attributes) below for details.

### Match Attributes

The `match_attributes` block supports the following arguments:

* `destination` - (Optional) Set of configuration blocks describing the destination IP address and address ranges to inspect for, in CIDR notation. If not specified, this matches with any destination address. See [Destination](#destination) below for details.

* `destination_port` - (Optional) Set of configuration blocks describing the destination ports to inspect for. If not specified, this matches with any destination port. See [Destination Port](#destination-port) below for details.

* `protocols` - (Optional) Set of protocols to inspect for, specified using the protocol's assigned internet protocol number (IANA). If not specified, this matches with any protocol.

* `source` - (Optional) Set of configuration blocks describing the source IP address and address ranges to inspect for, in CIDR notation. If not specified, this matches with any source address. See [Source](#source) below for details.

* `source_port` - (Optional) Set of configuration blocks describing the source ports to inspect for. If not specified, this matches with any source port. See [Source Port](#source-port) below for details.

* `tcp_flag` - (Optional) Set of configuration blocks containing the TCP flags and masks to inspect for. If not specified, this matches with any settings.

### Action Definition

The `action_definition` block supports the following argument:

* `publish_metric_action` - (Required) A configuration block describing the stateless inspection criteria that publishes the specified metrics to Amazon CloudWatch for the matching packet. You can pair this custom action with any of the standard stateless rule actions. See [Publish Metric Action](#publish-metric-action) below for details.

### Publish Metric Action

The `publish_metric_action` block supports the following argument:

* `dimension` - (Required) Set of configuration blocks containing the dimension settings to use for Amazon CloudWatch custom metrics. See [Dimension](#dimension) below for details.

### Dimension

The `dimension` block supports the following argument:

* `value` - (Required) The value to use in the custom metric dimension.

### Destination

The `destination` block supports the following argument:

* `address_definition` - (Required)  An IP address or a block of IP addresses in CIDR notation. AWS Network Firewall supports all address ranges for IPv4.

### Destination Port

The `destination_port` block supports the following arguments:

* `from_port` - (Required) The lower limit of the port range. This must be less than or equal to the `to_port`.

* `to_port` - (Optional) The upper limit of the port range. This must be greater than or equal to the `from_port`.

### Source

The `source` block supports the following argument:

* `address_definition` - (Required)  An IP address or a block of IP addresses in CIDR notation. AWS Network Firewall supports all address ranges for IPv4.

### Source Port

The `source_port` block supports the following arguments:

* `from_port` - (Required) The lower limit of the port range. This must be less than or equal to the `to_port`.

* `to_port` - (Optional) The upper limit of the port range. This must be greater than or equal to the `from_port`.

### TCP Flag

The `tcp_flag` block supports the following arguments:

* `flags` - (Required) Set of flags to look for in a packet. This setting can only specify values that are also specified in `masks`.
Valid values: `FIN`, `SYN`, `RST`, `PSH`, `ACK`, `URG`, `ECE`, `CWR`.

* `masks` - (Optional) Set of flags to consider in the inspection. To inspect all flags, leave this empty.
Valid values: `FIN`, `SYN`, `RST`, `PSH`, `ACK`, `URG`, `ECE`, `CWR`.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - The Amazon Resource Name (ARN) that identifies the rule group.

* `arn` - The Amazon Resource Name (ARN) that identifies the rule group.

* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

* `update_token` - A string token used when updating the rule group.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Network Firewall Rule Groups using their `arn`. For example:

```terraform
import {
  to = aws_networkfirewall_rule_group.example
  id = "arn:aws:network-firewall:us-west-1:123456789012:stateful-rulegroup/example"
}
```

Using `terraform import`, import Network Firewall Rule Groups using their `arn`. For example:

```console
% terraform import aws_networkfirewall_rule_group.example arn:aws:network-firewall:us-west-1:123456789012:stateful-rulegroup/example
```
