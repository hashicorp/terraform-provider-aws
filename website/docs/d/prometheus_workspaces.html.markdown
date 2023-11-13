---
subcategory: "AMP (Managed Prometheus)"
layout: "aws"
page_title: "AWS: aws_prometheus_workspaces"
description: |-
  Gets the aliases, ARNs, and workspace IDs of Amazon Prometheus workspaces.
---

# Data Source: aws_prometheus_workspaces

Provides the aliases, ARNs, and workspace IDs of Amazon Prometheus workspaces.

## Example Usage

The following example returns all of the workspaces in a region:

```terraform
data "aws_prometheus_workspaces" "example" {}
```

The following example filters the workspaces by alias. Only the workspaces with
aliases that begin with the value of `alias_prefix` will be returned:

```terraform
data "aws_prometheus_workspaces" "example" {
  alias_prefix = "example"
}
```

## Argument Reference

This data source supports the following arguments:

* `alias_prefix` - (Optional) Limits results to workspaces with aliases that begin with this value.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `aliases` - List of aliases of the matched Prometheus workspaces.
* `arns` - List of ARNs of the matched Prometheus workspaces.
* `workspace_ids` - List of workspace IDs of the matched Prometheus workspaces.
