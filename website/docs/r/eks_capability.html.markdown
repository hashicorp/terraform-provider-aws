---
subcategory: "EKS (Elastic Kubernetes)"
layout: "aws"
page_title: "AWS: aws_eks_capability"
description: |-
  Manages an EKS Capability.
---

# Resource: aws_eks_capability

Manages an EKS Capability for an EKS cluster.

## Example Usage

### Basic Usage

```terraform
resource "aws_eks_capability" "example" {
  cluster_name              = aws_eks_cluster.example.name
  capability_name           = "argocd"
  type                      = "ARGOCD"
  role_arn                  = aws_iam_role.example.arn
  delete_propagation_policy = "RETAIN"

  configuration {
    argo_cd {
      aws_idc {
        idc_instance_arn = "arn:aws:sso:::instance/ssoins-1234567890abcdef0"
      }
      namespace = "argocd"
    }
  }

  tags = {
    Name = "example-capability"
  }
}
```

### Controller Log Delivery

Capability controllers run in AWS-managed infrastructure outside your cluster, and their logs are exposed through [CloudWatch Vended Logs](https://docs.aws.amazon.com/eks/latest/userguide/capabilities-controller-logs.html) rather than the EKS API. Configure delivery with the [`aws_cloudwatch_log_delivery_source`](/docs/providers/aws/r/cloudwatch_log_delivery_source.html), [`aws_cloudwatch_log_delivery_destination`](/docs/providers/aws/r/cloudwatch_log_delivery_destination.html), and [`aws_cloudwatch_log_delivery`](/docs/providers/aws/r/cloudwatch_log_delivery.html) resources, using the capability ARN as the source. Valid log types are `EKS_CAPABILITY_ACK_LOGS` (ACK), `EKS_CAPABILITY_KRO_LOGS` (kro), and `EKS_CAPABILITY_ARGOCD_APPLICATION_LOGS`, `EKS_CAPABILITY_ARGOCD_APPLICATIONSET_LOGS`, `EKS_CAPABILITY_ARGOCD_COMMITSERVER_LOGS`, `EKS_CAPABILITY_ARGOCD_REPOSERVER_LOGS`, and `EKS_CAPABILITY_ARGOCD_SERVER_LOGS` (Argo CD, one per controller component).

```terraform
resource "aws_eks_capability" "example" {
  cluster_name              = aws_eks_cluster.example.name
  capability_name           = "ack"
  type                      = "ACK"
  role_arn                  = aws_iam_role.example.arn
  delete_propagation_policy = "RETAIN"
}

resource "aws_cloudwatch_log_group" "ack" {
  name = "/aws/eks/example/capabilities/ack"
}

resource "aws_cloudwatch_log_delivery_source" "ack" {
  name         = "eks-capability-ack-logs"
  log_type     = "EKS_CAPABILITY_ACK_LOGS"
  resource_arn = aws_eks_capability.example.arn
}

resource "aws_cloudwatch_log_delivery_destination" "ack" {
  name = "eks-capability-ack-logs"

  delivery_destination_configuration {
    destination_resource_arn = aws_cloudwatch_log_group.ack.arn
  }
}

resource "aws_cloudwatch_log_delivery" "ack" {
  delivery_source_name     = aws_cloudwatch_log_delivery_source.ack.name
  delivery_destination_arn = aws_cloudwatch_log_delivery_destination.ack.arn
}
```

## Argument Reference

This resource supports the following arguments:

* `capability_name` - (Required) Name of the capability. Must be unique within the cluster.
* `cluster_name` - (Required) Name of the EKS cluster.
* `configuration` - (Optional) Configuration for the capability. See [`configuration`](#configuration) below.
* `delete_propagation_policy` - (Required) Delete propagation policy for the capability. Valid values: `RETAIN`.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `role_arn` - (Required) ARN of the IAM role to associate with the capability.
* `tags` - (Optional) Key-value map of resource tags.
* `type` - (Required) Type of the capability. Valid values: `ACK`, `KRO`, `ARGOCD`.

### `configuration`

The `configuration` block contains the following:

* `argo_cd` - (Optional) ArgoCD configuration. See [`argo_cd`](#argo_cd) below.

### `argo_cd`

The `argo_cd` block contains the following:

* `aws_idc` - (Optional) AWS IAM Identity Center configuration. See [`aws_idc`](#aws_idc) below.
* `namespace` - (Optional) Kubernetes namespace for ArgoCD.
* `network_access` - (Optional) Network access configuration. See [`network_access`](#network_access) below.
* `rbac_role_mapping` - (Optional) RBAC role mappings. See [`rbac_role_mapping`](#rbac_role_mapping) below.

### `aws_idc`

The `aws_idc` block contains the following:

* `idc_instance_arn` - (Required) ARN of the IAM Identity Center instance.
* `idc_region` - (Optional) Region of the IAM Identity Center instance.

### `network_access`

The `network_access` block contains the following:

* `vpce_ids` - (Optional) VPC Endpoint IDs.

### `rbac_role_mapping`

The `rbac_role_mapping` block contains the following:

* `identity` - (Required) List of identities. See [`identity`](#identity) below.
* `role` - (Required) ArgoCD role. Valid values: `ADMIN`, `EDITOR`, `VIEWER`.

### `identity`

The `identity` block contains the following:

* `id` - (Required) Identity ID.
* `type` - (Required) Identity type. Valid values: `SSO_USER`, `SSO_GROUP`.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the capability.
* `configuration.0.argo_cd.0.server_url` - URL of the Argo CD server.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).
* `version` - Version of the capability.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `20m`)
* `update` - (Default `20m`)
* `delete` - (Default `20m`)

## Import

In Terraform v1.12.0 and later, the [`import` block](https://developer.hashicorp.com/terraform/language/import) can be used with the `identity` attribute. For example:

```terraform
import {
  to = aws_eks_capability.example
  identity = {
    cluster_name    = "example-cluster"
    capability_name = "example-capability"
  }
}

resource "aws_eks_capability" "example" {
  ### Configuration omitted for brevity ###
}
```

### Identity Schema

#### Required

* `cluster_name` (String) Name of the EKS Cluster.
* `capability_name` (String) Name of the capability.

#### Optional

* `account_id` (String) AWS Account where this resource is managed.
* `region` (String) Region where this resource is managed.

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Capabilities using `cluster_name` and `capability_name` separated by a comma (`,`). For example:

```terraform
import {
  to = aws_eks_capability.example
  id = "example-cluster,example-capability"
}
```

Using `terraform import`, import Capabilities using `cluster_name` and `capability_name` separated by a comma (`,`). For example:

```console
% terraform import aws_eks_capability.example example-cluster,example-capability
```
