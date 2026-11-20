---
subcategory: "CloudWatch Omni"
layout: "aws"
page_title: "AWS: aws_cloudwatchomni_organization_access_grant"
description: |-
  Manages a CloudWatch Omni Organization Access Grant.
---

# Resource: aws_cloudwatchomni_organization_access_grant

Manages a CloudWatch Omni Organization Access Grant. An organization access grant gives a principal a permission across an [organization domain](cloudwatchomni_organization_domain.html).

~> **NOTE:** This resource must be managed from the organization's management account, in the organization domain's home Region. Trusted access for the `observabilityadmin.amazonaws.com` service principal must be enabled in AWS Organizations.

~> **NOTE:** CloudWatch Omni does not support updating an organization access grant, so changing any argument other than `tags` replaces it. Tags are updated in place.

~> **NOTE:** A principal can hold only one grant of each permission in an organization domain. Creating a duplicate returns `ConflictException`.

~> **NOTE:** CloudWatch Omni creates a service-managed grant for the creator alongside each organization domain. It cannot be deleted and is not managed by this resource.

## Example Usage

### Basic Usage

```terraform
resource "aws_cloudwatchomni_organization_access_grant" "example" {
  name       = "example"
  domain_id  = aws_cloudwatchomni_organization_domain.example.domain_id
  permission = "ADMIN"

  principal {
    principal_type = "IAM_ROLE"
    principal_id   = aws_iam_role.example.arn
  }

  tags = {
    Environment = "production"
  }
}
```

## Argument Reference

The following arguments are required:

* `domain_id` - (Required) ID of the [CloudWatch Omni organization domain](cloudwatchomni_organization_domain.html) the grant applies to. Changing this forces a new resource.
* `name` - (Required) Name of the access grant. Must be 1-64 characters and contain only letters, numbers, underscores (`_`) and hyphens (`-`). Changing this forces a new resource.
* `permission` - (Required) Permission to grant. Valid value is `ADMIN`. Changing this forces a new resource.
* `principal` - (Required) Principal receiving the grant. [See below](#principal-block). Changing this forces a new resource.

The following arguments are optional:

* `region` - (Optional) Region where this resource is managed. Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `tags` - (Optional) Map of tags assigned to the resource. If configured with a provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

### `principal` Block

At least one of `principal_id` or `principal_attribute` must be specified.

* `principal_attribute` - (Optional) Attribute condition for attribute-based access. When set, the grant targets any principal matching all conditions. Supported only when `principal_type` is `IDC_USER`. Up to 10 blocks. [See below](#principal_attribute-block).
* `principal_id` - (Optional) ID of the principal, such as an IAM role ARN or an Identity Center user or group ID. Must be 1-2048 characters and start with a letter or number.
* `principal_type` - (Required) Type of principal. Valid values are `IDC_USER`, `IDC_GROUP`, `IAM_USER`, `IAM_ROLE` and `IAM_ROOT`.

### `principal_attribute` Block

* `key` - (Required) Identity Center user attribute to match, for example `userName` or `enterprise.department`. Must be 1-128 characters. Each key may appear only once per grant.
* `value` - (Required) Attribute value. Must be 1-512 characters.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the organization access grant.
* `grant_id` - Unique ID of the organization access grant.
* `grant_type` - Who manages the organization access grant.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Import

In Terraform v1.12.0 and later, you can use an [`import` block](https://developer.hashicorp.com/terraform/language/import) with the `identity` attribute. For example:

```terraform
import {
  to = aws_cloudwatchomni_organization_access_grant.example
  identity = {
    grant_id = "12345678-1234-1234-1234-123456789012"
  }
}

resource "aws_cloudwatchomni_organization_access_grant" "example" {
  ### Configuration omitted for brevity ###
}
```

### Identity Schema

#### Required

* `grant_id` (String) Organization access grant ID.

#### Optional

* `account_id` (String) AWS Account where this resource is managed.
* `region` (String) Region where this resource is managed.

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import a CloudWatch Omni Organization Access Grant using the `grant_id`. For example:

```terraform
import {
  to = aws_cloudwatchomni_organization_access_grant.example
  id = "12345678-1234-1234-1234-123456789012"
}
```

Using `terraform import`, import a CloudWatch Omni Organization Access Grant using the `grant_id`. For example:

```console
% terraform import aws_cloudwatchomni_organization_access_grant.example 12345678-1234-1234-1234-123456789012
```
