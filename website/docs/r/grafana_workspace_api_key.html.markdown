---
subcategory: "Managed Grafana"
layout: "aws"
page_title: "AWS: aws_grafana_workspace_api_key"
description: |-
  Creates a Grafana API key for the workspace. This key can be used to authenticate requests sent to the workspace's HTTP API.
---

# Resource: aws_grafana_workspace_api_key

Provides an Amazon Managed Grafana workspace API Key resource.

## Example Usage

### Basic configuration

```terraform
resource "aws_grafana_workspace_api_key" "key" {
  key_name        = "test-key"
  key_role        = "VIEWER"
  seconds_to_live = 3600
  workspace_id    = aws_grafana_workspace.test.id
}
```

## Argument Reference

The following arguments are required:

- `key_name` - (Required) Specifies the name of the API key. Key names must be unique to the workspace.
- `key_role` - (Required) Specifies the permission level of the API key. Valid values are `VIEWER`, `EDITOR`, or `ADMIN`.
- `seconds_to_live` - (Required) Specifies the time in seconds until the API key expires. Keys can be valid for up to 30 days.
- `workspace_id` - (Required) The ID of the workspace that the API key is valid for.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `key` - The key token in JSON format. Use this value as a bearer token to authenticate HTTP requests to the workspace.
