---
subcategory: "Network Security Manager"
layout: "aws"
page_title: "AWS: aws_networksecuritymanager_deployment"
description: |-
  Terraform resource for managing an AWS Network Security Manager Deployment.
---

# Resource: aws_networksecuritymanager_deployment

Terraform resource for managing an AWS Network Security Manager Deployment. A deployment applies one or two published policies to the resources selected by one published scope. See [Deployments](https://docs.aws.amazon.com/network-security-manager/latest/devguide/concepts.html) in the AWS Network Security Manager Developer Guide.

~> **A published deployment changes real resources.** Once a deployment is published (`is_published = true`, the default), the service evaluates every resource of the scope's resource types and, for each in-scope resource and each policy whose `remediation_enabled` is `true`, creates and associates the protections the policy describes: an AWS WAF web ACL named `NSMManagedWebACL-<policy names>-<timestamp>` for a `WAF` policy, an AWS Shield Advanced protection named `NSMManagedShieldProtection-<resource id>` for a `SHIELD_ADVANCED` policy. Remediation is asynchronous and has been observed to take from under a minute to more than ten minutes; this resource does not wait for it. A policy with `remediation_enabled = false` only reports the resources as out of sync. Whether the protections are removed again when the deployment is deleted is decided by each policy's `resources_clean_up` setting, not by this resource: with `resources_clean_up = true` the web ACL or protection is removed within seconds of the deployment's deletion; with `resources_clean_up = false` it stays behind after `terraform destroy`, still associated with the resource. Protections are not removed when a resource merely leaves the scope. Review the scope before publishing a deployment, and prefer scopes that select resources explicitly or by tag.

~> **Destroy the deployment before its policies.** After `DeleteDeployment` the service first disassociates the web ACL it created and deletes it a few seconds later; a policy deleted in between leaves the web ACL behind, unassociated, and since the service marks its web ACLs as managed by Firewall Manager the account can no longer delete it (`AccessDeniedException: … managed by Firewall Manager`). Terraform destroys a deployment before the policies it references, but does not wait for the service's clean-up. When destroying a deployment together with its policies, remove the deployment in a first apply and confirm the `NSMManagedWebACL-*` web ACL is gone (`aws wafv2 list-web-acls --scope REGIONAL`) before destroying the policies.

~> A resource that already carries a web ACL of its own is handled according to the WAF policy's `existing_customer_web_acl_resolution`: with `NO_REMEDIATION` it is left alone and reported out of sync; with `RETROFIT` the policy's rule groups are added to the existing web ACL as Firewall Manager rule groups, and they stay there after the deployment is deleted; with `OVERRIDE_ASSOCIATION` the resource is re-associated with the web ACL the service creates, and after the deployment is deleted the resource is left without any web ACL, the original one is not re-attached.

~> A `WAF` policy can only be remediated when its rules together describe a complete web ACL, at least a `DefaultAction` and a `VisibilityConfig` configuration rule. A policy that lacks one of them is reported out of sync but never creates a web ACL, without an error.

A deployment that exists only as a draft (`is_published = false` and never published) applies nothing and does not protect its scope or policies from deletion. Only published (`ACTIVE`) scopes and policies can be referenced.

~> A scope or policy referenced by a published deployment cannot be deleted. When such a scope or policy must be replaced (for example after a change of its `name`), set `create_before_destroy = true` in its `lifecycle` block so that the deployment is moved to the new resource before the old one is deleted; otherwise the apply fails with a `ConflictException`.

## Example Usage

### Basic Usage

```terraform
resource "aws_networksecuritymanager_rule" "default_action" {
  name          = "default-action-allow"
  firewall_type = "WAF"
  rule_type     = "CONFIGURATION"
  configuration = jsonencode({
    DefaultAction = {
      Allow = {}
    }
  })
}

resource "aws_networksecuritymanager_policy" "example" {
  name          = "web-acl-baseline"
  firewall_type = "WAF"
  priority      = 1

  policy_configuration {
    remediation_enabled = true
    resources_clean_up  = true

    waf_config {
      conflict_resolution                  = "MERGE_WHERE_APPLICABLE"
      existing_customer_web_acl_resolution = "NO_REMEDIATION"
    }
  }

  associated_template_and_rule {
    rule_arn = aws_networksecuritymanager_rule.default_action.arn
  }
}

resource "aws_networksecuritymanager_scope" "example" {
  name = "tagged-load-balancers"

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

resource "aws_networksecuritymanager_deployment" "example" {
  name        = "web-acl-baseline"
  policy_arns = [aws_networksecuritymanager_policy.example.arn]
  scope_arn   = aws_networksecuritymanager_scope.example.arn

  deployment_configuration {
    enable_cross_account_visibility = false
  }
}
```

### WAF and Shield Advanced Policies

```terraform
resource "aws_networksecuritymanager_deployment" "example" {
  name        = "edge-protection"
  description = "Web ACL and DDoS protection for every tagged distribution"
  policy_arns = [
    aws_networksecuritymanager_policy.waf.arn,
    aws_networksecuritymanager_policy.shield_advanced.arn,
  ]
  scope_arn = aws_networksecuritymanager_scope.distributions.arn

  deployment_configuration {
    enable_cross_account_visibility = true
  }
}
```

### Draft Deployment

```terraform
resource "aws_networksecuritymanager_deployment" "example" {
  name         = "web-acl-baseline"
  is_published = false
  policy_arns  = [aws_networksecuritymanager_policy.example.arn]
  scope_arn    = aws_networksecuritymanager_scope.example.arn

  deployment_configuration {
    enable_cross_account_visibility = false
  }
}
```

### Replacing a Referenced Scope

```terraform
resource "aws_networksecuritymanager_scope" "example" {
  name = "tagged-load-balancers"

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

  lifecycle {
    create_before_destroy = true
  }
}

resource "aws_networksecuritymanager_deployment" "example" {
  name        = "web-acl-baseline"
  policy_arns = [aws_networksecuritymanager_policy.example.arn]
  scope_arn   = aws_networksecuritymanager_scope.example.arn

  deployment_configuration {
    enable_cross_account_visibility = false
  }
}
```

## Argument Reference

The following arguments are required:

* `deployment_configuration` - (Required) Settings of the deployment. See [`deployment_configuration` Block](#deployment_configuration-block) below.
* `name` - (Required, Forces new resource) Name of the deployment. Must start with an alphanumeric character and contain only alphanumeric characters, spaces, and `_.:/=+-@`. Deployment names are not unique; a deployment is identified by its `arn`.
* `policy_arns` - (Required) Set of ARNs of the published policies the deployment applies, between 1 and 2.
* `scope_arn` - (Required) ARN of the published scope that selects the resources the deployment protects.

The following arguments are optional:

* `description` - (Optional) Description of the deployment, up to 256 characters.
* `is_published` - (Optional) Whether the deployment is published. When `true` (the default) the deployment is `ACTIVE` and its policies are applied to the resources in its scope. When `false` the deployment is saved as a `DRAFT` and applies nothing; setting it to `false` on a published deployment saves a draft on top of the published version, which stays in force until the draft is published.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#region).
* `tags` - (Optional) Map of tags assigned to the resource. If configured with a provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

### `deployment_configuration` Block

The `deployment_configuration` block supports the following arguments:

* `enable_cross_account_visibility` - (Required) Whether the aggregate synchronization status of the resources covered by the deployment is visible across the accounts of the organization.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the deployment.
* `coverage` - Coverage of the deployment, one entry per firewall type. See [`coverage` Block](#coverage-block) below.
* `deployment_id` - Service-generated ID of the deployment.
* `has_published_version` - Whether a published version exists beneath a pending draft.
* `status` - Status of the deployment. Valid values: `DRAFT`, `ACTIVE`, `DISABLED`.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).
* `updated_at` - Time when the deployment was last updated.
* `version` - Version of the deployment. Incremented on every update.
* `warnings` - Warnings about the deployment, such as a policy with no applicable resource in the scope. See [`warnings` Block](#warnings-block) below.

### `coverage` Block

The `coverage` block exports the following attributes:

* `firewall_type` - Firewall type shared by the policies of the entry.
* `in_scope_resource_types` - Resource types selected by the scope that the firewall type protects. Empty when the scope selects no resource type the firewall type can protect.
* `policy_arns` - ARNs of the deployment's policies of this firewall type.

### `warnings` Block

The `warnings` block exports the following attributes:

* `code` - Code identifying the kind of warning.
* `message` - Description of the warning.
* `policy_arn` - ARN of the policy the warning relates to.

## Import

In Terraform v1.12.0 and later, the [`import` block](https://developer.hashicorp.com/terraform/language/import) can be used with the `identity` attribute. For example:

```terraform
import {
  to = aws_networksecuritymanager_deployment.example
  identity = {
    arn = "arn:aws:network-security-manager:us-east-1:123456789012:deployment:1uas6q72iip9haw26eu596f7t"
  }
}
```

### Identity Schema

#### Required

* `arn` - (String) ARN of the deployment.

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Network Security Manager Deployment using the `arn`. For example:

```terraform
import {
  to = aws_networksecuritymanager_deployment.example
  id = "arn:aws:network-security-manager:us-east-1:123456789012:deployment:1uas6q72iip9haw26eu596f7t"
}
```

Using `terraform import`, import Network Security Manager Deployment using the `arn`. For example:

```console
% terraform import aws_networksecuritymanager_deployment.example arn:aws:network-security-manager:us-east-1:123456789012:deployment:1uas6q72iip9haw26eu596f7t
```
