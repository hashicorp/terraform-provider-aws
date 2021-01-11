---
subcategory: "EKS"
layout: "aws"
page_title: "AWS: aws_eks_addon"
description: |-
  Retrieve information about an EKS Addon
---

# Data Source: aws_eks_addon

Retrieve information about an EKS Addon.

## Example Usage

```hcl
data "aws_eks_addon" "example" {
  addon_name   = "vpc-cni"
  cluster_name = aws_eks_cluster.example.name
}

output "eks_addon_outputs" {
  value = aws_eks_addon.example
}
```

## Argument Reference

* `addon_name` – (Required) Name of the EKS Addon. The name must match one of
  the names returned by [list-addon](https://docs.aws.amazon.com/cli/latest/reference/eks/list-addons.html).
* `cluster_name` – (Required) Name of the EKS Cluster.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - Amazon Resource Name (ARN) of the EKS Addon.
* `addon_version` - The version of EKS Addon.
* `service_account_role_arn` - ARN of IAM role used for EKS Addon. If value is empty -
  then Addon uses the IAM role assigned to the EKS Cluster node.
* `id` - EKS Cluster name and EKS Addon name separated by a colon (`:`).
* `status` - Status of the EKS Addon.
* `created_at` - Date and time in [RFC3339 format](https://tools.ietf.org/html/rfc3339#section-5.8) that the EKS Addon was created.
* `modified_at` - Date and time in [RFC3339 format](https://tools.ietf.org/html/rfc3339#section-5.8) that the EKS Addon was updated.
