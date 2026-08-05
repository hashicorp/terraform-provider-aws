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

* `action` - (Required) Action for the listener rule. See [`action` Block](#action-block) for details.
* `listener_identifier` - (Required) ID or Amazon Resource Name (ARN) of the listener.
* `match` - (Required) Rule match. See [`match` Block](#match-block) for details.
* `name` - (Required) Name of the rule. Must be unique within the listener. Valid characters are a-z, 0-9, and hyphens (-). You can't use a hyphen as the first or last character, or immediately after another hyphen.
* `priority` - (Required) Priority assigned to the rule. Each rule for a specific listener must have a unique priority. The lower the priority number the higher the priority.
* `service_identifier` - (Required) ID or Amazon Resource Name (ARN) of the service.

The following arguments are optional:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `tags` - (Optional) Key-value mapping of resource tags. If configured with a provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

### `action` Block

The `action` block supports the following:

Exactly one of `fixed_response` or `forward` is required.

* `fixed_response` - (Optional) Rule action that returns a custom HTTP response. See [`fixed_response` Block](#fixed_response-block) for details.
* `forward` - (Optional) Forward action. Traffic that matches the rule is forwarded to the specified target groups. See [`forward` Block](#forward-block) for details.

### `fixed_response` Block

The `fixed_response` block supports the following:

* `status_code` - (Optional) HTTP response code.

### `forward` Block

The `forward` block supports the following:

* `target_groups` - (Required) Target groups that traffic matching the rule is forwarded to. See [`target_groups` Block](#target_groups-block) for details.

### `target_groups` Block

The `target_groups` block supports the following:

* `target_group_identifier` - (Required) ID or ARN of the target group.
* `weight` - (Optional) Weight assigned to the target group, controlling the prioritization and selection of each target group so that requests are distributed based on their weights. Default is `100`.

### `match` Block

The `match` block supports the following:

* `http_match` - (Required) HTTP criteria that a rule must match. See [`http_match` Block](#http_match-block) for details.

### `http_match` Block

The `http_match` block supports the following:

At least one of `header_matches`, `method`, or `path_match` is required.

* `header_matches` - (Optional) Header matches that match incoming requests based on the request header value before applying the rule action. See [`header_matches` Block](#header_matches-block) for details.
* `method` - (Optional) HTTP method type.
* `path_match` - (Optional) Path match. See [`path_match` Block](#path_match-block) for details.

### `header_matches` Block

The `header_matches` block supports the following:

* `case_sensitive` - (Optional) Whether the match is case sensitive. Default is `false`.
* `match` - (Optional) Header match type. See [`match.http_match.header_matches.match` Block](#matchhttp_matchheader_matchesmatch-block) for details.
* `name` - (Required) Name of the header.

### `match.http_match.header_matches.match` Block

The `match.http_match.header_matches.match` block supports the following:

Exactly one of `contains`, `exact`, or `prefix` is required.

* `contains` - (Optional) Contains type match.
* `exact` - (Optional) Exact type match.
* `prefix` - (Optional) Prefix type match. Matches the value with the prefix.

### `path_match` Block

The `path_match` block supports the following:

* `case_sensitive` - (Optional) Whether the match is case sensitive. Default is `false`.
* `match` - (Optional) Path match type. See [`match.http_match.path_match.match` Block](#matchhttp_matchpath_matchmatch-block) for details.

### `match.http_match.path_match.match` Block

The `match.http_match.path_match.match` block supports the following:

Exactly one of `exact` or `prefix` is required.

* `exact` - (Optional) Exact type match.
* `prefix` - (Optional) Prefix type match. Matches the value with the prefix.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN for the listener rule.
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
