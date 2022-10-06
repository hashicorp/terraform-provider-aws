---
subcategory: "Managed Grafana"
layout: "aws"
page_title: "AWS: aws_grafana_license_association"
description: |-
  Provides an Amazon Managed Grafana workspace license association resource.
---

# Resource: aws_grafana_license_association

Provides an Amazon Managed Grafana workspace license association resource.

## Example Usage

### Basic configuration

```terraform
resource "aws_grafana_license_association" "example" {
  license_type = "ENTERPRISE_FREE_TRIAL"
  workspace_id = aws_grafana_workspace.example.id
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

* `license_type` - (Required) The type of license for the workspace license association. Valid values are `ENTERPRISE` and `ENTERPRISE_FREE_TRIAL`.
* `workspace_id` - (Required) The workspace id.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `free_trial_expiration` - If `license_type` is set to `ENTERPRISE_FREE_TRIAL`, this is the expiration date of the free trial.
* `license_expiration` - If `license_type` is set to `ENTERPRISE`, this is the expiration date of the enterprise license.

## Import

Grafana workspace license association can be imported using the workspace's `id`, e.g.,

```
$ terraform import aws_grafana_license_association.example g-2054c75a02
```
