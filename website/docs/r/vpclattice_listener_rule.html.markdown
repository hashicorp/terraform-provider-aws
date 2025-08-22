---
subcategory: "VPC Lattice"
layout: "aws"
page_title: "AWS: aws_vpclattice_listener_rule"
description: |-
  Terraform resource for managing an AWS VPC Lattice Listener Rule.
---

# Resource: aws_vpclattice_listener_rule

Terraform resource for managing an AWS VPC Lattice Listener Rule.

## Example Usage

```terraform
resource "aws_vpclattice_listener_rule" "example" {
  name                = "example"
  listener_identifier = aws_vpclattice_listener.example.listener_id
  service_identifier  = aws_vpclattice_service.example.id
  priority            = 20
  match {
    http_match {

      header_matches {
        name           = "example-header"
        case_sensitive = false

        match {
          exact = "example-contains"
        }
      }

      path_match {
        case_sensitive = true
        match {
          prefix = "/example-path"
        }
      }
    }
  }
  action {
    forward {
      target_groups {
        target_group_identifier = aws_vpclattice_target_group.example.id
        weight                  = 1
      }
      target_groups {
        target_group_identifier = aws_vpclattice_target_group.example2.id
        weight                  = 2
      }
    }

  }
}
```

### Basic Usage

```terraform
resource "aws_vpclattice_listener_rule" "example" {
  name                = "example"
  listener_identifier = aws_vpclattice_listener.example.listener_id
  service_identifier  = aws_vpclattice_service.example.id
  priority            = 10
  match {
    http_match {
      path_match {
        case_sensitive = false
        match {
          exact = "/example-path"
        }
      }
    }
  }
  action {
    fixed_response {
      status_code = 404
    }
  }
}
```

## Argument Reference

The following arguments are required:

* `service_identifier` - (Required) The ID or Amazon Resource Identifier (ARN) of the service.
* `listener_identifier` - (Required) The ID or Amazon Resource Name (ARN) of the listener.
* `action` - (Required) The action for the listener rule.
  See [`action` Block](#action-block) for details.
* `match` - (Required) The rule match.
  See [`match` Block](#match-block)
* `name` - (Required) The name of the rule. The name must be unique within the listener. The valid characters are a-z, 0-9, and hyphens (-). You can't use a hyphen as the first or last character, or immediately after another hyphen.
* `priority` - (Required) The priority assigned to the rule. Each rule for a specific listener must have a unique priority. The lower the priority number the higher the priority.

The following arguments are optional:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `tags` - (Optional) Key-value mapping of resource tags. If configured with a provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

### `action` Block

The `action` block supports the following:

Exactly one of `fixed_response` or `forward` is required.

* `fixed_response` - (Optional) Describes the rule action that returns a custom HTTP response.
  See [`fixed_response` Block](#fixed_response-block) for details.
* `forward` - (Optional) The forward action. Traffic that matches the rule is forwarded to the specified target groups.
  See [`forward` Block](#forward-block) for details.

### `fixed_response` Block

The `fixed_response` block supports the following:

* `status_code` - (Optional) The HTTP response code.

### `forward` Block

The `forward` block supports the following:

* `target_groups` - (Optional) The target groups. Traffic matching the rule is forwarded to the specified target groups. With forward actions, you can assign a weight that controls the prioritization and selection of each target group. This means that requests are distributed to individual target groups based on their weights. For example, if two target groups have the same weight, each target group receives half of the traffic.

The default value is 1 with maximum number of 2. If only one target group is provided, there is no need to set the weight; 100% of traffic will go to that target group.

### `match` Block

The `match` block supports the following:

* `http_match` - (Required) The HTTP criteria that a rule must match.
  See [`http_match` Block](#http_match-block) for details.

### `http_match` Block

The `http_match` block supports the following:

At least one of `header_matches`, `method`, or `path_match` is required.

* `header_matches` - (Optional) The header matches.
  Matches incoming requests with rule based on request header value before applying rule action.
  See [`header_matches` Block](#header_matches-block) for details.
* `method` - (Optional) The HTTP method type.
* `path_match` - (Optional) The path match.
  See [`path_match` Block](#path_match-block) for details.

### `header_matches` Block

The `header_matches` block supports the following:

* `case_sensitive` - (Optional) Indicates whether the match is case sensitive.
  Default is `false`.
* `match` - (Optional) The header match type.
  See [Header Match `match` Block](#header-match-match-block) for details.
* `name` - (Required) The name of the header.

### Header Match `match` Block

The Header Match `match` block supports the following:

Exactly one of `contains`, `exact`, or `prefix` is required.

* `contains` - (Optional) Specifies a contains type match.
* `exact` - (Optional) Specifies an exact type match.
* `prefix` - (Optional) Specifies a prefix type match.
  Matches the value with the prefix.

### `path_match` Block

The `path_match` block supports the following:

* `case_sensitive` - (Optional) Indicates whether the match is case sensitive.
  Default is `false`.
* `match` - (Optional) The header match type.
  See [Path Match `match` Block](#path-match-match-block) for details.

### Path Match `match` Block

The Path Match `match` block supports the following:

Exactly one of `exact` or `prefix` is required.

* `exact` - (Optional) Specifies an exact type match.
* `prefix` - (Optional) Specifies a prefix type match.
  Matches the value with the prefix.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - The ARN for the listener rule.
* `rule_id` - Unique identifier for the listener rule.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block).

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `60m`)
* `update` - (Default `180m`)
* `delete` - (Default `90m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import VPC Lattice Listener Rule using the `id`. For example:

```terraform
import {
  to = aws_vpclattice_listener_rule.example
  id = "service123/listener456/rule789"
}
```

Using `terraform import`, import VPC Lattice Listener Rule using the `id`. For example:

```console
% terraform import aws_vpclattice_listener_rule.example service123/listener456/rule789
```
