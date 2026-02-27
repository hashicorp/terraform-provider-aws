---
subcategory: "WorkMail"
layout: "aws"
page_title: "AWS: aws_workmail_organization"
description: |-
  Manages an AWS WorkMail Organization.
---

# Resource: aws_workmail_organization

Manages an AWS WorkMail Organization.

## Example Usage

### Basic Usage

```terraform
resource "aws_workmail_organization" "example" {
  organization_alias = "example-org"
}
```

## Argument Reference

The following arguments are required:

* `organization_alias` - (Required) Alias for the organization. Must be unique globally. Changing this creates a new resource.

The following arguments are optional:

* `delete_directory` - (Optional) Whether to delete the IAM Identity Center directory associated with the organization on destroy. To update this value after creation, run `terraform apply` before running `terraform destroy`. Defaults to `false`.
* `directory_id` - (Optional) ID of an existing directory to associate with the organization. Changing this creates a new resource.
* `interoperability_enabled` - (Optional) Whether to enable interoperability between WorkMail and Microsoft Exchange. Changing this creates a new resource.
* `kms_key_arn` - (Optional) ARN of a customer-managed KMS key to encrypt the organization's data. If omitted, AWS managed keys are used. Changing this creates a new resource.
* `tags` - (Optional) A map of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the Organization.
* `completed_date` - Date and time (RFC3339) at which the organization became active.
* `default_mail_domain` - Default mail domain for the organization.
* `directory_type` - Type of the associated directory.
* `migration_admin` - User ID of the migration admin if migration is enabled.
* `organization_id` - ID of the WorkMail Organization.
* `state` - State of the organization.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `30m`)
* `delete` - (Default `30m`)

## Import

In Terraform v1.12.0 and later, the [`import` block](https://developer.hashicorp.com/terraform/language/import) can be used with the `identity` attribute. For example:

```terraform
import {
  to = aws_workmail_organization.example
  identity = {
    organization_id = "m-xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"
  }
}

resource "aws_workmail_organization" "example" {
  ### Configuration omitted for brevity ###
}
```

### Identity Schema

#### Required

* `organization_id` - (String) ID of the WorkMail Organization.

#### Optional

* `account_id` (String) AWS Account where this resource is managed.
* `region` (String) Region where this resource is managed.

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import WorkMail Organization using the `organization_id`. For example:

```terraform
import {
  to = aws_workmail_organization.example
  id = "m-xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"
}
```

Using `terraform import`, import WorkMail Organization using the `organization_id`. For example:

```console
% terraform import aws_workmail_organization.example m-xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
```

~> **NOTE:** The `kms_key_arn` and `delete_directory` attributes are not returned by the AWS API and will not be set after import. Add them back to your configuration manually if needed.
