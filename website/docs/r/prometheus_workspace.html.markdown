---
subcategory: "Amazon Managed Service for Prometheus (AMP)"
layout: "aws"
page_title: "AWS: aws_prometheus_workspace"
description: |-
  Manages an Amazon Managed Service for Prometheus (AMP) Workspace
---

# Resource: aws_prometheus_workspace

Manages an Amazon Managed Service for Prometheus (AMP) Workspace.

~> **NOTE:** This AWS functionality is in Preview and may change before General Availability release. Backwards compatibility is not guaranteed between Terraform AWS Provider releases.

## Example Usage

```terraform
resource "aws_prometheus_workspace" "demo" {
  alias = "prometheus-test"

  tags = {
    Environment = "production"
    Owner       = "abhi"
  }
}
```

## Argument Reference

The following argument is supported:

* `alias` - (Optional) The alias of the prometheus workspace. See more [in AWS Docs](https://docs.aws.amazon.com/prometheus/latest/userguide/AMP-onboard-create-workspace.html).
* `tags` - (Optional) A map of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - Amazon Resource Name (ARN) of the workspace.
* `id` - Identifier of the workspace
* `prometheus_endpoint` - Prometheus endpoint available for this workspace.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block).

## Import

AMP Workspaces can be imported using the identifier, e.g.,

```
$ terraform import aws_prometheus_workspace.demo ws-C6DCB907-F2D7-4D96-957B-66691F865D8B
```
