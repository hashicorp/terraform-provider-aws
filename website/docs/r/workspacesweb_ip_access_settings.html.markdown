---
subcategory: "WorkSpaces Web"
layout: "aws"
page_title: "AWS: aws_workspacesweb_ip_access_settings"
description: |-
  Terraform resource for managing an AWS WorkSpaces Web IP Access Settings.
---

# Resource: aws_workspacesweb_ip_access_settings

Terraform resource for managing an AWS WorkSpaces Web IP Access Settings resource. Once associated with a web portal, IP access settings control which IP addresses users can connect from.

## Example Usage

### Basic Usage

```terraform
resource "aws_workspacesweb_ip_access_settings" "example" {
  display_name = "example"
  ip_rule {
    ip_range = "10.0.0.0/16"
  }
}
```

### With Multiple IP Rules

```terraform
resource "aws_workspacesweb_ip_access_settings" "example" {
  display_name = "example"
  description  = "Example IP access settings"
  ip_rule {
    ip_range    = "10.0.0.0/16"
    description = "Main office"
  }
  ip_rule {
    ip_range    = "192.168.0.0/24"
    description = "Branch office"
  }
}
```

### With All Arguments

```terraform
resource "aws_kms_key" "example" {
  description             = "KMS key for WorkSpaces Web IP Access Settings"
  deletion_window_in_days = 7
}

resource "aws_workspacesweb_ip_access_settings" "example" {
  display_name         = "example"
  description          = "Example IP access settings"
  customer_managed_key = aws_kms_key.example.arn
  additional_encryption_context = {
    Environment = "Production"
  }
  ip_rule {
    ip_range    = "10.0.0.0/16"
    description = "Main office"
  }
  ip_rule {
    ip_range    = "192.168.0.0/24"
    description = "Branch office"
  }
  tags = {
    Name = "example-ip-access-settings"
  }
}
```

## Argument Reference

The following arguments are required:

* `display_name` - (Required) The display name of the IP access settings.
* `ip_rule` - (Required) The IP rules of the IP access settings. See [IP Rule](#ip-rules) below.

The following arguments are optional:

* `additional_encryption_context` - (Optional) Additional encryption context for the IP access settings.
* `customer_managed_key` - (Optional) ARN of the customer managed KMS key.
* `description` - (Optional) The description of the IP access settings.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `tags` - (Optional) Map of tags assigned to the resource. If configured with a provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

### IP Rules

* `ip_range` - (Required) The IP range of the IP rule.
* `description` - (Optional) The description of the IP rule.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `associated_portal_arns` - List of web portal ARNs that this IP access settings resource is associated with.
* `ip_access_settings_arn` - ARN of the IP access settings resource.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block).

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import WorkSpaces Web IP Access Settings using the `ip_access_settings_arn`. For example:

```terraform
import {
  to = aws_workspacesweb_ip_access_settings.example
  id = "arn:aws:workspaces-web:us-west-2:123456789012:ipAccessSettings/abcdef12345"
}
```

Using `terraform import`, import WorkSpaces Web IP Access Settings using the `ip_access_settings_arn`. For example:

```console
% terraform import aws_workspacesweb_ip_access_settings.example arn:aws:workspaces-web:us-west-2:123456789012:ipAccessSettings/abcdef12345
```
