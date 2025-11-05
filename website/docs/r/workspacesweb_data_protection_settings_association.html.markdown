---
subcategory: "WorkSpaces Web"
layout: "aws"
page_title: "AWS: aws_workspacesweb_data_protection_settings_association"
description: |-
  Terraform resource for managing an AWS WorkSpaces Web Data Protection Settings Association.
---

# Resource: aws_workspacesweb_data_protection_settings_association

Terraform resource for managing an AWS WorkSpaces Web Data Protection Settings Association.

## Example Usage

### Basic Usage

```terraform
resource "aws_workspacesweb_portal" "example" {
  display_name = "example"
}

resource "aws_workspacesweb_data_protection_settings" "example" {
  display_name = "example"
}

resource "aws_workspacesweb_data_protection_settings_association" "example" {
  data_protection_settings_arn = aws_workspacesweb_data_protection_settings.example.data_protection_settings_arn
  portal_arn                   = aws_workspacesweb_portal.example.portal_arn
}
```

## Argument Reference

The following arguments are required:

* `data_protection_settings_arn` - (Required) ARN of the data protection settings to associate with the portal. Forces replacement if changed.
* `portal_arn` - (Required) ARN of the portal to associate with the data protection settings. Forces replacement if changed.

The following arguments are optional:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

## Attribute Reference

This resource exports no additional attributes.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import WorkSpaces Web Data Protection Settings Association using the `data_protection_settings_arn,portal_arn`. For example:

```terraform
import {
  to = aws_workspacesweb_data_protection_settings_association.example
  id = "arn:aws:workspaces-web:us-west-2:123456789012:dataProtectionSettings/data_protection_settings-id-12345678,arn:aws:workspaces-web:us-west-2:123456789012:portal/portal-id-12345678"
}
```
