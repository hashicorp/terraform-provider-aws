---
subcategory: "Managed Grafana"
layout: "aws"
page_title: "AWS: aws_grafana_workspaces"
description: |-
  Gets the names and workspace IDs of Amazon Managed Grafana workspaces.
---

# Data Source: aws_grafana_workspaces

Provides the names and workspace IDs of Amazon Managed Grafana workspaces.

## Example Usage

The following example returns all of the workspaces in a region:

```terraform
data "aws_grafana_workspaces" "example" {}
```

The following example filters the workspaces by name. Only the workspaces with
a name that matches the provided value of `name` will be returned:

```terraform
data "aws_grafana_workspaces" "example" {
  name = "example"
}
```

## Argument Reference

This data source supports the following arguments:

* `name` - (Optional) Limits results to workspaces with a name that matches with this value.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `names` - List of aliases of the matched Grafana workspaces.
* `workspace_ids` - List of workspace IDs of the matched Grafana workspaces.
