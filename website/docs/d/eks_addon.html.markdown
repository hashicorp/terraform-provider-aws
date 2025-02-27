---
subcategory: "EKS (Elastic Kubernetes)"
layout: "aws"
page_title: "AWS: aws_eks_addon"
description: |-
  Retrieve information about an EKS add-on
---

# Data Source: aws_eks_addon

Retrieve information about an EKS add-on.

## Example Usage

```terraform
data "aws_eks_addon" "example" {
  addon_name   = "vpc-cni"
  cluster_name = aws_eks_cluster.example.name
}

output "eks_addon_outputs" {
  value = aws_eks_addon.example
}
```

## Argument Reference

* `addon_name` – (Required) Name of the EKS add-on. The name must match one of
  the names returned by [list-addon](https://docs.aws.amazon.com/cli/latest/reference/eks/list-addons.html).
* `cluster_name` – (Required) Name of the EKS Cluster.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `arn` - ARN of the EKS add-on.
* `addon_version` - Version of EKS add-on.
* `configuration_values` - Configuration values for the addon with a single JSON string.
* `service_account_role_arn` - ARN of IAM role used for EKS add-on. If value is empty -
  then add-on uses the IAM role assigned to the EKS Cluster node.
* `pod_identity_association` - Pod identity association for the EKS add-on.
    * `role_arn` - ARN of the IAM role associated with the EKS add-on.
    * `service_account` - Service account associated with the EKS add-on.
* `id` - EKS Cluster name and EKS add-on name separated by a colon (`:`).
* `created_at` - Date and time in [RFC3339 format](https://tools.ietf.org/html/rfc3339#section-5.8) that the EKS add-on was created.
* `modified_at` - Date and time in [RFC3339 format](https://tools.ietf.org/html/rfc3339#section-5.8) that the EKS add-on was updated.
