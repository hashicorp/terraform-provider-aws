---
subcategory: "Managed Grafana"
layout: "aws"
page_title: "AWS: aws_grafana_workspace_saml_configuration"
description: |-
  Provides an Amazon Managed Grafana workspace SAML configuration resource.
---

# Resource: aws_grafana_workspace_saml_configuration

Provides an Amazon Managed Grafana workspace SAML configuration resource.

## Example Usage

### Basic configuration

```terraform
resource "aws_grafana_workspace_saml_configuration" "example" {
  editor_role_values = ["editor"]
  idp_metadata_url   = "https://my_idp_metadata.url"
  workspace_id       = aws_grafana_workspace.example.id
}

resource "aws_grafana_workspace" "example" {
  account_access_type      = "CURRENT_ACCOUNT"
  authentication_providers = ["SAML"]
  permission_type          = "SERVICE_MANAGED"
  role_arn                 = aws_iam_role.assume.arn
}

resource "aws_iam_role" "assume" {
  name = "grafana-assume"
  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRole"
        Effect = "Allow"
        Sid    = ""
        Principal = {
          Service = "grafana.amazonaws.com"
        }
      },
    ]
  })
}
```

## Argument Reference

The following arguments are required:

* `editor_role_values` - (Required) The editor role values.
* `workspace_id` - (Required) The workspace id.

The following arguments are optional:

* `admin_role_values` - (Optional) The admin role values.
* `allowed_organizations` - (Optional) The allowed organizations.
* `email_assertion` - (Optional) The email assertion.
* `groups_assertion` - (Optional) The groups assertion.
* `idp_metadata_url` - (Optional) The IDP Metadata URL. Note that either `idp_metadata_url` or `idp_metadata_xml` (but not both) must be specified.
* `idp_metadata_xml` - (Optional) The IDP Metadata XML. Note that either `idp_metadata_url` or `idp_metadata_xml` (but not both) must be specified.
* `login_assertion` - (Optional) The login assertion.
* `login_validity_duration` - (Optional) The login validity duration.
* `name_assertion` - (Optional) The name assertion.
* `org_assertion` - (Optional) The org assertion.
* `role_assertion` - (Optional) The role assertion.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `status` - The status of the SAML configuration.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Grafana Workspace SAML configuration using the workspace's `id`. For example:

```terraform
import {
  to = aws_grafana_workspace_saml_configuration.example
  id = "g-2054c75a02"
}
```

Using `terraform import`, import Grafana Workspace SAML configuration using the workspace's `id`. For example:

```console
% terraform import aws_grafana_workspace_saml_configuration.example g-2054c75a02
```
