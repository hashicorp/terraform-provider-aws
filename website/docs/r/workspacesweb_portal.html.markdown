---
subcategory: "WorkSpaces Web"
layout: "aws"
page_title: "AWS: aws_workspacesweb_portal"
description: |-
  Terraform resource for managing an AWS WorkSpaces Web Portal.
---

# Resource: aws_workspacesweb_portal

Terraform resource for managing an AWS WorkSpaces Web Portal.

## Example Usage

### Basic Usage

```terraform
resource "aws_workspacesweb_portal" "example" {
  display_name  = "example-portal"
  instance_type = "standard.regular"
}
```

### Complete Usage

```terraform
resource "aws_kms_key" "example" {
  description             = "KMS key for WorkSpaces Web Portal"
  deletion_window_in_days = 7
}

resource "aws_workspacesweb_portal" "example" {
  display_name            = "example-portal"
  instance_type           = "standard.large"
  authentication_type     = "IAM_Identity_Center"
  customer_managed_key    = aws_kms_key.example.arn
  max_concurrent_sessions = 10

  additional_encryption_context = {
    Environment = "Production"
  }

  tags = {
    Name = "example-portal"
  }

  timeouts {
    create = "10m"
    update = "10m"
    delete = "10m"
  }
}
```

## Argument Reference

The following arguments are optional:

* `additional_encryption_context` - (Optional) Additional encryption context for the customer managed key. Forces replacement if changed.
* `authentication_type` - (Optional) Authentication type for the portal. Valid values: `Standard`, `IAM_Identity_Center`.
* `browser_settings_arn` - (Optional) ARN of the browser settings to use for the portal.
* `customer_managed_key` - (Optional) ARN of the customer managed key. Forces replacement if changed.
* `display_name` - (Optional) Display name of the portal.
* `instance_type` - (Optional) Instance type for the portal. Valid values: `standard.regular`, `standard.large`.
* `max_concurrent_sessions` - (Optional) Maximum number of concurrent sessions for the portal.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `tags` - (Optional) Key-value map of resource tags. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `browser_type` - Browser type of the portal.
* `creation_date` - Creation date of the portal.
* `data_protection_settings_arn` - ARN of the data protection settings associated with the portal.
* `ip_access_settings_arn` - ARN of the IP access settings associated with the portal.
* `network_settings_arn` - ARN of the network settings associated with the portal.
* `portal_arn` - ARN of the portal.
* `portal_endpoint` - Endpoint URL of the portal.
* `portal_status` - Status of the portal.
* `renderer_type` - Renderer type of the portal.
* `session_logger_arn` - ARN of the session logger associated with the portal.
* `status_reason` - Reason for the current status of the portal.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).
* `trust_store_arn` - ARN of the trust store associated with the portal.
* `user_access_logging_settings_arn` - ARN of the user access logging settings associated with the portal.
* `user_settings_arn` - ARN of the user settings associated with the portal.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `5m`)
* `update` - (Default `5m`)
* `delete` - (Default `5m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import WorkSpaces Web Portal using the `portal_arn`. For example:

```terraform
import {
  to = aws_workspacesweb_portal.example
  id = "arn:aws:workspaces-web:us-west-2:123456789012:portal/abcdef12345678"
}
```

Using `terraform import`, import WorkSpaces Web Portal using the `portal_arn`. For example:

```console
% terraform import aws_workspacesweb_portal.example arn:aws:workspaces-web:us-west-2:123456789012:portal/abcdef12345678
```
