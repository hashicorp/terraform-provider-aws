---
subcategory: "Network Security Manager"
layout: "aws"
page_title: "AWS: aws_networksecuritymanager_scope"
description: |-
  Terraform resource for managing an AWS Network Security Manager Scope.
---

# Resource: aws_networksecuritymanager_scope

Terraform resource for managing an AWS Network Security Manager Scope. A scope selects the accounts and the resources that a deployment applies policies to; the scope itself deploys nothing. See [Scopes](https://docs.aws.amazon.com/network-security-manager/latest/devguide/concepts.html) in the AWS Network Security Manager Developer Guide.

A scope selects resources per resource type with one `resource_scope` block per type. Each block either includes every resource of the type (`include_all = true`), or includes or excludes the resources named by explicit ARNs and/or matched by a logical expression over tag criteria and, for Application Load Balancers, load balancer configuration criteria. An expression is a single `criteria`, or a single `and`, `or` or `not` operator over criteria: the API does not accept an operator nested under another operator.

CloudFront distributions are a global resource type and every other type is regional. A scope selects either global or regional resource types, never both, and cannot change between them once created: such a change replaces the scope.

An `account_filter` block is required when the account is the administrator of an AWS Organizations organization, and is not accepted when the account is a single-account administrator: a scope without an account filter applies to the administrator's own account. Whether a scope has an account filter is fixed when it is created; adding or removing one replaces the scope.

## Example Usage

### Tagged Distributions

```terraform
resource "aws_networksecuritymanager_scope" "example" {
  name = "production-distributions"

  scope_configuration {
    resource_scope {
      resource_type = "AWS::CloudFront::Distribution"

      include {
        expression {
          criteria {
            tags = {
              environment = "production"
            }
          }
        }
      }
    }
  }
}
```

### Internet-Facing Load Balancers Except Some

```terraform
resource "aws_networksecuritymanager_scope" "example" {
  name        = "public-load-balancers"
  description = "Internet-facing load balancers not marked as exempt"

  scope_configuration {
    resource_scope {
      resource_type = "AWS::ElasticLoadBalancingV2::LoadBalancer::application"

      include {
        explicit_arns = [aws_lb.example.arn]

        expression {
          and {
            criteria {
              alb_config {
                scheme = "internet-facing"
              }
            }
            criteria {
              tags = {
                environment = "production"
              }
            }
          }
        }
      }
    }

    resource_scope {
      resource_type = "AWS::ApiGateway::Stage"

      exclude {
        expression {
          criteria {
            tags = {
              waf-exempt = "true"
            }
          }
        }
      }
    }
  }
}
```

### Organization Scope

```terraform
resource "aws_networksecuritymanager_scope" "example" {
  name = "workloads"

  scope_configuration {
    account_filter {
      include {
        organizational_units = [aws_organizations_organizational_unit.workloads.id]
      }
    }

    resource_scope {
      resource_type = "AWS::CloudFront::Distribution"
      include_all   = true
    }
  }
}
```

## Argument Reference

The following arguments are required:

* `name` - (Required, Forces new resource) Name of the scope. Must start with an alphanumeric character and contain only alphanumeric characters, spaces, and `_.:/=+-@`. Scope names are not unique; a scope is identified by its `arn`.
* `scope_configuration` - (Required) Accounts and resources the scope selects. See [`scope_configuration` Block](#scope_configuration-block) below.

The following arguments are optional:

* `description` - (Optional) Description of the scope, up to 256 characters.
* `is_published` - (Optional) Whether the scope is published. When `true` (the default) the scope is `ACTIVE` and can be used by deployments. When `false` the scope is saved as a `DRAFT`; setting it to `false` on a published scope saves a draft on top of the published version, which stays in use until the draft is published.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#region).
* `tags` - (Optional) Map of tags assigned to the resource. If configured with a provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

### `scope_configuration` Block

The `scope_configuration` block supports the following arguments:

* `account_filter` - (Optional, Forces new resource when added or removed) Accounts the scope selects. Required for an organization administrator, not accepted for a single-account administrator. See [`account_filter` Block](#account_filter-block) below.
* `resource_scope` - (Required) Resources the scope selects, one block per resource type. All blocks must be of global resource types or all of regional resource types. See [`resource_scope` Block](#resource_scope-block) below.

### `account_filter` Block

Exactly one of the arguments of the `account_filter` block must be set:

* `exclude` - (Optional) Accounts and organizational units to exclude; every other account of the organization is selected. See [`exclude` Block](#exclude-block) below.
* `include` - (Optional) Accounts and organizational units to select. See [`include` Block](#include-block) below.
* `include_all` - (Optional) Set to `true` to select every account of the organization.

### `exclude` Block

The `exclude` block of an `account_filter` block supports the following arguments:

* `account_ids` - (Optional) Set of up to 1000 AWS account IDs.
* `organizational_units` - (Optional) Set of up to 1000 organizational unit IDs (`ou-...`).

### `include` Block

The `include` block of an `account_filter` block supports the following arguments:

* `account_ids` - (Optional) Set of up to 1000 AWS account IDs.
* `organizational_units` - (Optional) Set of up to 1000 organizational unit IDs (`ou-...`).

### `resource_scope` Block

Exactly one of `exclude`, `include` and `include_all` must be set in a `resource_scope` block:

* `exclude` - (Optional) Resources of the type to exclude; every other resource of the type is selected. See [`resource_scope.exclude` Block](#resource_scopeexclude-block) below.
* `include` - (Optional) Resources of the type to select. See [`resource_scope.include` Block](#resource_scopeinclude-block) below.
* `include_all` - (Optional) Set to `true` to select every resource of the type.
* `resource_type` - (Required) Resource type. Valid values: `AWS::ApiGateway::Stage`, `AWS::CloudFront::Distribution`, `AWS::EC2::EIP`, `AWS::ElasticLoadBalancing::LoadBalancer`, `AWS::ElasticLoadBalancingV2::LoadBalancer::application`. `AWS::CloudFront::Distribution` is the only global resource type.

### `resource_scope.exclude` Block

The `exclude` block of a `resource_scope` block supports the following arguments:

* `explicit_arns` - (Optional) Set of up to 100 resource ARNs. The ARNs are not checked against existing resources.
* `expression` - (Optional) Logical expression that matches resources. See [`expression` Block](#expression-block) below.

### `resource_scope.include` Block

The `include` block of a `resource_scope` block supports the following arguments:

* `explicit_arns` - (Optional) Set of up to 100 resource ARNs. The ARNs are not checked against existing resources.
* `expression` - (Optional) Logical expression that matches resources. See [`expression` Block](#expression-block) below.

### `expression` Block

Exactly one of the arguments of the `expression` block must be set. The API does not accept an operator under another operator, so the operands of `and`, `or` and `not` are all `criteria` blocks:

* `and` - (Optional) Resources matching every one of between 1 and 20 `criteria` blocks. See [`and` Block](#and-block) below.
* `criteria` - (Optional) Resources matching one criteria. See [`criteria` Block](#criteria-block) below.
* `not` - (Optional) Resources not matching a single `criteria` block. See [`not` Block](#not-block) below.
* `or` - (Optional) Resources matching at least one of between 1 and 20 `criteria` blocks. See [`or` Block](#or-block) below.

### `and` Block

* `criteria` - (Required) Between 1 and 20 criteria that must all match. See [`criteria` Block](#criteria-block) below.

### `not` Block

* `criteria` - (Required) Single criteria that must not match. See [`criteria` Block](#criteria-block) below.

### `or` Block

* `criteria` - (Required) Between 1 and 20 criteria of which at least one must match. See [`criteria` Block](#criteria-block) below.

### `criteria` Block

Exactly one of the arguments of a `criteria` block must be set:

* `alb_config` - (Optional) Application Load Balancer configuration to match. Only meaningful for the `AWS::ElasticLoadBalancingV2::LoadBalancer::application` resource type. See [`alb_config` Block](#alb_config-block) below.
* `tags` - (Optional) Map of up to 100 tag keys and values a resource must carry. An empty map matches every resource.

### `alb_config` Block

* `ip_address_type` - (Optional) IP address type of the load balancer. Valid values: `ipv4`, `dualstack`, `dualstack-without-public-ipv4`.
* `scheme` - (Optional) Scheme of the load balancer. Valid values: `internet-facing`, `internal`.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the scope.
* `has_published_version` - Whether a published version exists beneath a pending draft.
* `scope_id` - Service-generated ID of the scope.
* `status` - Status of the scope. Valid values: `DRAFT`, `ACTIVE`, `DISABLED`.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).
* `updated_at` - Time when the scope was last updated.
* `version` - Version of the scope. Incremented on every update.

## Import

In Terraform v1.12.0 and later, the [`import` block](https://developer.hashicorp.com/terraform/language/import) can be used with the `identity` attribute. For example:

```terraform
import {
  to = aws_networksecuritymanager_scope.example
  identity = {
    arn = "arn:aws:network-security-manager:us-east-1:123456789012:scope:1uas4tnr31knytfql818l5gss"
  }
}
```

### Identity Schema

#### Required

* `arn` - (String) ARN of the scope.

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Network Security Manager Scope using the `arn`. For example:

```terraform
import {
  to = aws_networksecuritymanager_scope.example
  id = "arn:aws:network-security-manager:us-east-1:123456789012:scope:1uas4tnr31knytfql818l5gss"
}
```

Using `terraform import`, import Network Security Manager Scope using the `arn`. For example:

```console
% terraform import aws_networksecuritymanager_scope.example arn:aws:network-security-manager:us-east-1:123456789012:scope:1uas4tnr31knytfql818l5gss
```
