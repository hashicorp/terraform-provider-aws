---
subcategory: "EKS (Elastic Kubernetes Service)"
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
  cluster_name                = aws_eks_cluster.example.name
  capability_name             = "argocd"
  type                        = "ARGOCD"
  role_arn                    = aws_iam_role.example.arn
  delete_propagation_policy   = "DELETE"

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
* `type` - (Required) Type of the capability (e.g., `ARGOCD`).
* `role_arn` - (Required) ARN of the IAM role to associate with the capability.
* `delete_propagation_policy` - (Required) Delete propagation policy for the capability. Valid values are `DELETE` and `ORPHAN`.
* `configuration` - (Optional) Configuration for the capability.
  * `argo_cd` - (Optional) ArgoCD configuration.
    * `aws_idc` - (Optional) AWS IAM Identity Center configuration.
      * `idc_instance_arn` - (Required) ARN of the IAM Identity Center instance.
      * `idc_region` - (Optional) Region of the IAM Identity Center instance.
    * `namespace` - (Optional) Kubernetes namespace for ArgoCD.
    * `network_access` - (Optional) Network access configuration.
      * `vpce_ids` - (Optional) VPC Endpoint IDs.
    * `rbac_role_mappings` - (Optional) RBAC role mappings.
      * `identities` - (Required) List of identities.
        * `id` - (Required) Identity ID.
        * `type` - (Required) Identity type.
      * `role` - (Required) ArgoCD role.
* `tags` - (Optional) Key-value map of resource tags.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the capability.
* `created_at` - Creation timestamp of the capability.
* `modified_at` - Last modification timestamp of the capability.
* `status` - Status of the capability.
* `version` - Version of the capability.

## Timeouts

`aws_eks_capability` provides the following [Timeouts](https://www.terraform.io/language/resources/syntax#timeouts) configuration options:

* `create` - (Default `20m`) How long to wait for the capability to be created.
* `update` - (Default `20m`) How long to wait for the capability to be updated.
* `delete` - (Default `20m`) How long to wait for the capability to be deleted.

## Import

EKS Capability can be imported using the `cluster_name` and `capability_name` separated by a colon (`:`), e.g.,

```
$ terraform import aws_eks_capability.example my-cluster:my-capability
```
