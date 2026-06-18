---
subcategory: "AMP (Managed Prometheus)"
layout: "aws"
page_title: "AWS: aws_prometheus_resource_policy"
description: |-
  Manages an Amazon Managed Service for Prometheus (AMP) Resource Policy.
---

# Resource: aws_prometheus_resource_policy

Manages an Amazon Managed Service for Prometheus (AMP) Resource Policy.

Resource-based policies allow you to grant permissions to other AWS accounts or services to access your Prometheus workspace. This enables cross-account access and fine-grained permissions for workspace sharing.

The following actions are supported in resource policies for Prometheus workspaces: `aps:RemoteWrite`, `aps:QueryMetrics`, `aps:GetSeries`, `aps:GetLabels`, `aps:GetMetricMetadata`.

~> **Note:** Only Prometheus-compatible APIs can be used for workspace sharing. Non-Prometheus-compatible APIs added to the policy will be ignored. If your workspace uses customer-managed KMS keys for encryption, you must grant the principals in your resource-based policy access to those KMS keys through KMS grants. The resource ARN in the policy document must match the workspace ARN that the policy is being attached to.

## Example Usage

### Basic Resource Policy

```terraform
resource "aws_prometheus_workspace" "example" {
  alias = "example-workspace"
}

data "aws_caller_identity" "current" {}

data "aws_iam_policy_document" "example" {
  statement {
    effect = "Allow"
    principals {
      type        = "AWS"
      identifiers = [data.aws_caller_identity.current.account_id]
    }
    actions = [
      "aps:RemoteWrite",
      "aps:QueryMetrics",
      "aps:GetSeries",
      "aps:GetLabels",
      "aps:GetMetricMetadata"
    ]
    resources = [aws_prometheus_workspace.example.arn]
  }
}

resource "aws_prometheus_resource_policy" "example" {
  workspace_id    = aws_prometheus_workspace.example.id
  policy_document = data.aws_iam_policy_document.example.json
}
```

### Cross-Account Access

```terraform
resource "aws_prometheus_workspace" "example" {
  alias = "example-workspace"
}

data "aws_iam_policy_document" "cross_account" {
  statement {
    effect = "Allow"
    principals {
      type        = "AWS"
      identifiers = ["arn:aws:iam::123456789012:root"]
    }
    actions = [
      "aps:RemoteWrite",
      "aps:QueryMetrics"
    ]
    resources = [aws_prometheus_workspace.example.arn]
  }
}

resource "aws_prometheus_resource_policy" "cross_account" {
  workspace_id    = aws_prometheus_workspace.example.id
  policy_document = data.aws_iam_policy_document.cross_account.json
}
```

### Service-Specific Access

```terraform
resource "aws_prometheus_workspace" "example" {
  alias = "example-workspace"
}

data "aws_iam_policy_document" "service_access" {
  statement {
    effect = "Allow"
    principals {
      type        = "Service"
      identifiers = ["grafana.amazonaws.com"]
    }
    actions = [
      "aps:QueryMetrics",
      "aps:GetSeries",
      "aps:GetLabels",
      "aps:GetMetricMetadata"
    ]
    resources = [aws_prometheus_workspace.example.arn]
  }
}

resource "aws_prometheus_resource_policy" "service_access" {
  workspace_id    = aws_prometheus_workspace.example.id
  policy_document = data.aws_iam_policy_document.service_access.json
}
```

## Argument Reference

This resource supports the following arguments:

* `workspace_id` - (Required) The ID of the workspace to attach the resource-based policy to.
* `policy_document` - (Required) The JSON policy document to use as the resource-based policy. This policy defines the permissions that other AWS accounts or services have to access your workspace.

The following arguments are optional:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `policy_status` - The current status of the resource-based policy. Can be `CREATING`, `ACTIVE`, `UPDATING`, or `DELETING`.
* `revision_id` - The revision ID of the current resource-based policy.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

- `create` - (Default `5m`)
- `update` - (Default `5m`)
- `delete` - (Default `5m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import the Resource Policy using the workspace ID. For example:

```terraform
import {
  to = aws_prometheus_resource_policy.example
  id = "ws-12345678-90ab-cdef-1234-567890abcdef"
}
```

Using `terraform import`, import AMP Resource Policies using the workspace ID. For example:

```console
% terraform import aws_prometheus_resource_policy.example ws-12345678-90ab-cdef-1234-567890abcdef
```
