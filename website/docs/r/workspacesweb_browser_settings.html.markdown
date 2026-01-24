---
subcategory: "WorkSpaces Web"
layout: "aws"
page_title: "AWS: aws_workspacesweb_browser_settings"
description: |-
  Terraform resource for managing an AWS WorkSpaces Web Browser Settings.
---

# Resource: aws_workspacesweb_browser_settings

Terraform resource for managing an AWS WorkSpaces Web Browser Settings resource.

## Example Usage

### Basic Usage

```terraform
resource "aws_workspacesweb_browser_settings" "example" {
  browser_policy = jsonencode({
    AdditionalSettings = {
      DownloadsSettings = {
        Behavior = "DISABLE"
      }
    }
  })
}
```

### With All Arguments

```terraform
resource "aws_kms_key" "example" {
  description             = "KMS key for WorkSpaces Web Browser Settings"
  deletion_window_in_days = 7
}

resource "aws_workspacesweb_browser_settings" "example" {
  browser_policy = jsonencode({
    chromePolicies = {
      DefaultDownloadDirectory = {
        value = "/home/as2-streaming-user/MyFiles/TemporaryFiles1"
      }
    }
  })
  customer_managed_key = aws_kms_key.example.arn
  additional_encryption_context = {
    Environment = "Production"
  }
  tags = {
    Name = "example-browser-settings"
  }
}
```

## Argument Reference

The following arguments are required:

* `browser_policy` - (Required) Browser policy for the browser settings. This is a JSON string that defines the browser settings policy.

The following arguments are optional:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `additional_encryption_context` - (Optional) Additional encryption context for the browser settings.
* `customer_managed_key` - (Optional) ARN of the customer managed KMS key.
* `tags` - (Optional) Map of tags assigned to the resource. If configured with a provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `associated_portal_arns` - List of web portal ARNs to associate with the browser settings.
* `browser_settings_arn` - ARN of the browser settings resource.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block).

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import WorkSpaces Web Browser Settings using the `browser_settings_arn`. For example:

```terraform
import {
  to = aws_workspacesweb_browser_settings.example
  id = "arn:aws:workspaces-web:us-west-2:123456789012:browsersettings/abcdef12345"
}
```

Using `terraform import`, import WorkSpaces Web Browser Settings using the `browser_settings_arn`. For example:

```console
% terraform import aws_workspacesweb_browser_settings.example arn:aws:workspacesweb:us-west-2:123456789012:browsersettings/abcdef12345
```
