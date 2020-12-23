---
subcategory: "Prometheus Service (AMP)"
layout: "aws"
page_title: "AWS: aws_prometheus_workspace"
description: |-
  Creates a prometheus workspace
---

# Resource: aws_prometheus_workspace

Provides a managed prometheus workspace.

-> **Note:** only workspace alias can be updated today

## Example Usage

```hcl
resource "aws_prometheus_workspace" "demo" {
  alias = "prometheus-test"
}
```

## Argument Reference

The following argument is supported:

* `alias` - (Optional) The alias of the prometheus workspace. See more [in AWS Docs](https://docs.aws.amazon.com/prometheus/latest/userguide/AMP-onboard-create-workspace.html).

## Attributes Reference

The following attribute is exported:

* `status` - The current status of the workspace. The possible values are CREATING, ACTIVE, UPDATING, DELETING, and CREATION_FAILED.
* `alias` - An alias that you assign to this workspace to help you identify it. It does not need to be unique.
* `prometheus_endpoint` - The Prometheus endpoint available for this workspace.


## Import

AMP Workspaces can be imported using the `arn`, e.g.

```
$ terraform import aws_prometheus_workspace.demo arn:aws:aps:us-west-2:123123123:workspace/ws-C6DCB907-F2D7-4D96-957B-66691F865D8B
```
