---
subcategory: "CloudWatch Omni"
layout: "aws"
page_title: "AWS: aws_cloudwatchomni_access_grant"
description: |-
  Manages a CloudWatch Omni Access Grant.
---

# Resource: aws_cloudwatchomni_access_grant

Manages a CloudWatch Omni Access Grant. An access grant gives a principal a permission within a CloudWatch Omni space.

~> **NOTE:** CloudWatch Omni does not support updating an access grant, so changing any argument replaces it.

~> **NOTE:** Except for `CUSTOM` grants, a principal can hold only one grant of each permission in a space. Creating a duplicate returns `ConflictException`.

~> **NOTE:** CloudWatch Omni creates service-managed access grants alongside each space. They cannot be deleted and are not managed by this resource.

## Example Usage

### Basic Usage

```terraform
resource "aws_cloudwatchomni_access_grant" "example" {
  name       = "example"
  domain_id  = aws_cloudwatchomni_domain.example.domain_id
  space_id   = aws_cloudwatchomni_space.example.space_id
  permission = "READ"

  principal {
    principal_type = "IAM_ROLE"
    principal_id   = aws_iam_role.example.arn
  }
}
```

### Custom Permission

A `CUSTOM` permission requires at least one `scoped_action` block.

```terraform
resource "aws_cloudwatchomni_access_grant" "example" {
  name       = "example"
  domain_id  = aws_cloudwatchomni_domain.example.domain_id
  space_id   = aws_cloudwatchomni_space.example.space_id
  permission = "CUSTOM"

  principal {
    principal_type = "IAM_ROLE"
    principal_id   = aws_iam_role.example.arn
  }

  scoped_action {
    actions = ["cloudwatch:GetRecords"]

    context_conditions = {
      channelNames = ["example"]
    }

    resource {
      resource_type = "DataSet"
      signal_types  = ["LOGS"]

      row_scope_group {
        row_scope {
          field    = "serviceName"
          operator = "IN"
          values   = ["checkout", "payments"]
        }
      }
    }
  }
}
```

## Argument Reference

The following arguments are required:

* `domain_id` - (Required) ID of the [CloudWatch Omni domain](cloudwatchomni_domain.html) that contains the space. Changing this forces a new resource.
* `name` - (Required) Name of the access grant. Changing this forces a new resource.
* `permission` - (Required) Permission to grant. Valid values are `SPACE_ADMIN`, `READ`, `READ_WRITE_DELETE` and `CUSTOM`. Changing this forces a new resource.
* `principal` - (Required) Principal receiving the grant. [See below](#principal-block). Changing this forces a new resource.
* `space_id` - (Required) ID of the [CloudWatch Omni space](cloudwatchomni_space.html) to scope the grant to. Changing this forces a new resource.

The following arguments are optional:

* `region` - (Optional) Region where this resource is managed. Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `scoped_action` - (Optional) Group of actions to allow, with the resource scopes and conditions that limit them. Used with the `CUSTOM` permission. [See below](#scoped_action-block). Changing this forces a new resource.

### `principal` Block

* `principal_attribute` - (Optional) Attribute condition for attribute-based access. When set, the grant targets any principal matching all conditions. Supported only for `IDC_USER` principals. [See below](#principal_attribute-block).
* `principal_id` - (Optional) ID of the principal, such as an IAM role ARN, an Identity Center user or group ID, or the ID of an access profile or alert.
* `principal_type` - (Required) Type of principal. Valid values are `IDC_USER`, `IDC_GROUP`, `IAM_USER`, `IAM_ROLE`, `IAM_ROOT`, `ACCESS_PROFILE`, `ALERT` and `AGENT`.

### `principal_attribute` Block

* `key` - (Required) Identity Center user attribute to match, for example `userName` or `enterprise.department`. Each key may appear only once per grant.
* `value` - (Required) Attribute value.

### `scoped_action` Block

* `actions` - (Required) Actions this group applies to, for example `cloudwatch:GetRecords`.
* `context_conditions` - (Optional) Map of context condition keys to allowed values, for fine-grained access control on these actions.
* `resource` - (Optional) Resource scope constraining these actions to specific resources. [See below](#resource-block).

### `resource` Block

* `resource_arns` - (Optional) Resource ARNs or ARN patterns the actions are limited to. Defaults to all resources of the type.
* `resource_type` - (Required) Resource type name, for example `DataSet` or `OmniDashboard`.
* `row_scope_group` - (Optional) Row-level filter group. A row is visible when it matches every `row_scope` in any one group. Requires `signal_types`. [See below](#row_scope_group-block).
* `signal_types` - (Optional) Signal types the row filtering applies to. Valid values are `LOGS` and `TRACES`.
* `tags` - (Optional) Map of tag key-value pairs. Access applies only to resources carrying all of these tags.

### `row_scope_group` Block

* `row_scope` - (Required) Row filter. [See below](#row_scope-block).

### `row_scope` Block

* `field` - (Required) Field the filter applies to, for example `serviceName`.
* `operator` - (Required) Match operator. Valid value is `IN`.
* `values` - (Required) Values the field is matched against.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the access grant.
* `created_at` - Time the access grant was created, in RFC3339 format.
* `grant_id` - Unique ID of the access grant.
* `grant_type` - Who manages the access grant.
* `updated_at` - Time the access grant was last updated, in RFC3339 format.

~> **NOTE:** This resource does not support tags. `CreateAccessGrant` accepts tags, but CloudWatch Omni does not return them on read and provides no tag read or write operations, so tags could not be refreshed, updated, or imported.

## Import

In Terraform v1.12.0 and later, you can use an [`import` block](https://developer.hashicorp.com/terraform/language/import) with the `identity` attribute. For example:

```terraform
import {
  to = aws_cloudwatchomni_access_grant.example
  identity = {
    grant_id = "12345678-1234-1234-1234-123456789012"
  }
}

resource "aws_cloudwatchomni_access_grant" "example" {
  ### Configuration omitted for brevity ###
}
```

### Identity Schema

#### Required

* `grant_id` (String) Access grant ID.

#### Optional

* `account_id` (String) AWS Account where this resource is managed.
* `region` (String) Region where this resource is managed.

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import a CloudWatch Omni Access Grant using the `grant_id`. For example:

```terraform
import {
  to = aws_cloudwatchomni_access_grant.example
  id = "12345678-1234-1234-1234-123456789012"
}
```

Using `terraform import`, import a CloudWatch Omni Access Grant using the `grant_id`. For example:

```console
% terraform import aws_cloudwatchomni_access_grant.example 12345678-1234-1234-1234-123456789012
```
