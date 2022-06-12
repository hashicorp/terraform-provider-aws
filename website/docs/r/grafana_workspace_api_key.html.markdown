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

* `editor_role_values` - (Required) The editor role values.
* `workspace_id` - (Required) The workspace id.



## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `key` - The key token in JSON format. Use this value as a bearer token to authenticate HTTP requests to the workspace.

## Import

Grafana Workspace API Key cannot be imported.