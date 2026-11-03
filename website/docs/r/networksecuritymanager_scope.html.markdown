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
* `scope_configuration` - (Required) Accounts and resources that are in scope. See [`scope_configuration` Block](#scope_configuration-block) below.

The following arguments are optional:

* `description` - (Optional) Description of the scope.
* `is_published` - (Optional) Whether to publish the scope. When `true`, the scope is saved in the published (`ACTIVE`) state. When `false`, it is saved as a draft (`DRAFT`); for a published scope, the draft is saved on top of the published version. Defaults to `true`.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#region).
* `tags` - (Optional) Map of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

### `scope_configuration` Block

* `account_filter` - (Optional) Accounts that are in scope. Organization administrators must set it, and single-account administrators must omit it. Adding or removing the block forces a new resource. See [`scope_configuration.account_filter` Block](#scope_configurationaccount_filter-block) below.
* `resource_scope` - (Required) Resources that are in scope, one block per resource type. A scope selects either the global resource type `AWS::CloudFront::Distribution` or regional resource types, not both. Changing between them forces a new resource. See [`scope_configuration.resource_scope` Block](#scope_configurationresource_scope-block) below.

### `scope_configuration.account_filter` Block

Exactly one of `exclude`, `include` and `include_all` must be set.

* `exclude` - (Optional) Accounts to exclude. All other accounts are in scope. See [`scope_configuration.account_filter.exclude` Block](#scope_configurationaccount_filterexclude-block) below.
* `include` - (Optional) Accounts to include. See [`scope_configuration.account_filter.include` Block](#scope_configurationaccount_filterinclude-block) below.
* `include_all` - (Optional) Set to `true` to include all accounts.

### `scope_configuration.account_filter.exclude` Block

* `account_ids` - (Optional) Set of AWS account IDs.
* `organizational_units` - (Optional) Set of AWS Organizations organizational unit IDs.

### `scope_configuration.account_filter.include` Block

* `account_ids` - (Optional) Set of AWS account IDs.
* `organizational_units` - (Optional) Set of AWS Organizations organizational unit IDs.

### `scope_configuration.resource_scope` Block

Exactly one of `exclude`, `include` and `include_all` must be set.

* `exclude` - (Optional) Resources to exclude. All other resources of the type are in scope. See [`scope_configuration.resource_scope.exclude` Block](#scope_configurationresource_scopeexclude-block) below.
* `include` - (Optional) Resources to include. See [`scope_configuration.resource_scope.include` Block](#scope_configurationresource_scopeinclude-block) below.
* `include_all` - (Optional) Set to `true` to include all resources of the type.
* `resource_type` - (Required) Resource type. Valid values: `AWS::ApiGateway::Stage`, `AWS::CloudFront::Distribution`, `AWS::EC2::EIP`, `AWS::ElasticLoadBalancing::LoadBalancer`, `AWS::ElasticLoadBalancingV2::LoadBalancer::application`.

### `scope_configuration.resource_scope.exclude` Block

* `explicit_arns` - (Optional) Set of resource ARNs.
* `expression` - (Optional) Expression that matches resources. See [`expression` Block](#expression-block) below.

### `scope_configuration.resource_scope.include` Block

* `explicit_arns` - (Optional) Set of resource ARNs.
* `expression` - (Optional) Expression that matches resources. See [`expression` Block](#expression-block) below.

### `expression` Block

Exactly one of `and`, `criteria`, `not` and `or` must be set. Each `and`, `not` and `or` block is one operand. Operators can't be nested.

* `and` - (Optional) Matches resources that match every operand. Up to 20 blocks. See [`and` Block](#and-block) below.
* `criteria` - (Optional) Matches resources that match the criteria. See [`criteria` Block](#criteria-block) below.
* `not` - (Optional) Matches resources that don't match the operand. Exactly 1 block. See [`not` Block](#not-block) below.
* `or` - (Optional) Matches resources that match at least one operand. Up to 20 blocks. See [`or` Block](#or-block) below.

### `and` Block

Exactly one of `alb_config` and `tags` must be set.

* `alb_config` - (Optional) Application Load Balancer properties to match. See [`alb_config` Block](#alb_config-block) below.
* `tags` - (Optional) Map of tags that a resource must have. An empty map matches every resource.

### `criteria` Block

Exactly one of `alb_config` and `tags` must be set.

* `alb_config` - (Optional) Application Load Balancer properties to match. See [`alb_config` Block](#alb_config-block) below.
* `tags` - (Optional) Map of tags that a resource must have. An empty map matches every resource.

### `not` Block

Exactly one of `alb_config` and `tags` must be set.

* `alb_config` - (Optional) Application Load Balancer properties to match. See [`alb_config` Block](#alb_config-block) below.
* `tags` - (Optional) Map of tags that a resource must have. An empty map matches every resource.

### `or` Block

Exactly one of `alb_config` and `tags` must be set.

* `alb_config` - (Optional) Application Load Balancer properties to match. See [`alb_config` Block](#alb_config-block) below.
* `tags` - (Optional) Map of tags that a resource must have. An empty map matches every resource.

### `alb_config` Block

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
