---
subcategory: "Managed Grafana"
layout: "aws"
page_title: "AWS: aws_grafana_workspace_api_key"
description: |-
  Provides an Amazon Managed Grafana Workspace Api Key resource.
---

# Resource: aws_grafana_workspace_api_key

Provides an Amazon Managed Grafana Workspace Api Key resource.

## Example Usage

### Basic configuration

```terraform
resource "aws_grafana_workspace_api_Key" "example" {
  workspace_id    = aws_grafana_workspace.example.id
  key_name        = "example-editor-key"
  key_role        = "EDITOR"
  seconds_to_live = 3600
}
```

## Argument Reference

The following arguments are required:

- `workspace_id` - (Required) The ID of the workspace that the key is valid for.
- `key_name` - (Required) Specifies the name of the key. Keynames must be unique to the workspace.
- `key_role` - (Required) Specifies the permission level of the key. Valid values are `VIEWER`, `EDITOR`, or `ADMIN`.
- `seconds_to_live` - (Required) Specifies the time in seconds until the key expires. Keys can be valid for up to 30 days.

## Attributes Reference

All the above arguments are also exported as attributes.

## Import

Grafana Workspace Api Key can be imported using the Workspace Api Key's `key_name`, e.g.,

```
$ terraform import aws_grafana_workspace_api_key.example example-editor-key
```
