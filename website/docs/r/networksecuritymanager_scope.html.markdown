---
subcategory: "Network Security Manager"
layout: "aws"
page_title: "AWS: aws_networksecuritymanager_scope"
description: |-
  Manages an AWS Network Security Manager Scope.
---

# Resource: aws_networksecuritymanager_scope

Manages an AWS Network Security Manager Scope. A scope selects the accounts and resources that a deployment applies policies to.

Changes to a scope are published unless `is_published` is `false`.

## Example Usage

### Basic Usage

```terraform
resource "aws_networksecuritymanager_scope" "example" {
  name        = "example"
  description = "Application Load Balancers tagged for protection"

  scope_configuration {
    resource_scope {
      resource_type = "AWS::ElasticLoadBalancingV2::LoadBalancer::application"

      include {
        expression {
          criteria {
            tags = {
              Protect = "true"
            }
          }
        }
      }
    }
  }
}
```

### Logical Expressions and Account Filter

```terraform
resource "aws_networksecuritymanager_scope" "example" {
  name = "example"

  scope_configuration {
    account_filter {
      include {
        organizational_units = ["ou-abcd-12345678"]
      }
    }

    resource_scope {
      resource_type = "AWS::ElasticLoadBalancingV2::LoadBalancer::application"

      include {
        expression {
          and {
            tags = {
              Protect = "true"
            }
          }
          and {
            alb_config {
              scheme = "internet-facing"
            }
          }
        }
      }
    }

    resource_scope {
      resource_type = "AWS::ApiGateway::Stage"
      include_all   = true
    }
  }
}
```

## Argument Reference

The following arguments are required:

* `name` - (Required, Forces new resource) Name of the scope.
* `scope_configuration` - (Required) Accounts and resources that are in scope. See [`scope_configuration`](#scope_configuration) below.

The following arguments are optional:

* `description` - (Optional) Description of the scope.
* `is_published` - (Optional) Whether to publish the scope. When `true`, the scope is saved in the published (`ACTIVE`) state. When `false`, it is saved as a draft (`DRAFT`); for a published scope, the draft is saved on top of the published version. Defaults to `true`.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#region).
* `tags` - (Optional) Map of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

### `scope_configuration`

* `account_filter` - (Optional) Accounts that are in scope. Organization administrators must set it, and single-account administrators must omit it. Adding or removing the block forces a new resource. See [`account_filter`](#account_filter) below.
* `resource_scope` - (Required) Resources that are in scope, one block per resource type. A scope selects either the global resource type `AWS::CloudFront::Distribution` or regional resource types, not both. Changing between them forces a new resource. See [`resource_scope`](#resource_scope) below.

### `account_filter`

Exactly one of the following must be set:

* `exclude` - (Optional) Accounts to exclude. All other accounts are in scope. See [`exclude` and `include` (account filter)](#exclude-and-include-account-filter) below.
* `include` - (Optional) Accounts to include. See [`exclude` and `include` (account filter)](#exclude-and-include-account-filter) below.
* `include_all` - (Optional) Set to `true` to include all accounts.

### `exclude` and `include` (account filter)

* `account_ids` - (Optional) Set of AWS account IDs.
* `organizational_units` - (Optional) Set of AWS Organizations organizational unit IDs.

### `resource_scope`

* `resource_type` - (Required) Resource type. Valid values: `AWS::ApiGateway::Stage`, `AWS::CloudFront::Distribution`, `AWS::EC2::EIP`, `AWS::ElasticLoadBalancing::LoadBalancer`, `AWS::ElasticLoadBalancingV2::LoadBalancer::application`.

Exactly one of the following must be set:

* `exclude` - (Optional) Resources to exclude. All other resources of the type are in scope. See [`exclude` and `include` (resource scope)](#exclude-and-include-resource-scope) below.
* `include` - (Optional) Resources to include. See [`exclude` and `include` (resource scope)](#exclude-and-include-resource-scope) below.
* `include_all` - (Optional) Set to `true` to include all resources of the type.

### `exclude` and `include` (resource scope)

* `explicit_arns` - (Optional) Set of resource ARNs.
* `expression` - (Optional) Expression that matches resources. See [`expression`](#expression) below.

### `expression`

Exactly one of the following must be set. Each `and`, `not` and `or` block is one operand, with the same arguments as [`criteria`](#criteria). Operators can't be nested.

* `and` - (Optional) Matches resources that match every operand. Up to 20 blocks.
* `criteria` - (Optional) Matches resources that match the criteria. See [`criteria`](#criteria) below.
* `not` - (Optional) Matches resources that don't match the operand. Exactly 1 block.
* `or` - (Optional) Matches resources that match at least one operand. Up to 20 blocks.

### `criteria`

Exactly one of the following must be set:

* `alb_config` - (Optional) Application Load Balancer properties to match. See [`alb_config`](#alb_config) below.
* `tags` - (Optional) Map of tags that a resource must have. An empty map matches every resource.

### `alb_config`

* `ip_address_type` - (Optional) IP address type. Valid values: `dualstack`, `dualstack-without-public-ipv4`, `ipv4`.
* `scheme` - (Optional) Scheme. Valid values: `internal`, `internet-facing`.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the scope.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Import

In Terraform v1.12.0 and later, the [`import` block](https://developer.hashicorp.com/terraform/language/import) can be used with the `identity` attribute. For example:

```terraform
import {
  to = aws_networksecuritymanager_scope.example
  identity = {
    arn = "arn:aws:network-security-manager:us-east-1:123456789012:scope:0123456789abcdef"
  }
}
```

### Identity Schema

#### Required

* `arn` (String) ARN of the scope.

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Network Security Manager Scope using the `arn`. For example:

```terraform
import {
  to = aws_networksecuritymanager_scope.example
  id = "arn:aws:network-security-manager:us-east-1:123456789012:scope:0123456789abcdef"
}
```

Using `terraform import`, import Network Security Manager Scope using the `arn`. For example:

```console
% terraform import aws_networksecuritymanager_scope.example arn:aws:network-security-manager:us-east-1:123456789012:scope:0123456789abcdef
```
