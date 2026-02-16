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

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import EKS Capability using the `cluster_name` and `capability_name` separated by a comma (`,`). For example:

```terraform
import {
  to = aws_eks_capability.example
  id = "my-cluster,my-capability"
}
```

Using `terraform import`, import EKS Capability using the `cluster_name` and `capability_name` separated by a comma (`,`). For example:

```console
% terraform import aws_eks_capability.example my-cluster,my-capability
```
