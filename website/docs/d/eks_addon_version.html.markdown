---
subcategory: "EKS"
layout: "aws"
page_title: "AWS: aws_eks_addon_version"
description: |-
  Retrieve information about versions of an EKS add-on
---

# Data Source: aws_eks_addon_version

Retrieve information about EKS add-on versions supported across EKS cluster versions.

## Example Usage

```terraform
data "aws_eks_addon_version" "default" {
  addon_name = "vpc-cni"
  kubernetes_version  = aws_eks_cluster.example.version
}

data "aws_eks_addon_version" "latest" {
  addon_name = "vpc-cni"
  kubernetes_version  = aws_eks_cluster.example.version
  latest = true
}

resource "aws_eks_addon" "vpc_cni" {
  cluster_name = aws_eks_cluster.example.name
  addon_name = "vpc-cni"
  addon_version =  data.aws_eks_addon_version.latest.version
}

output "default" {
  value = data.aws_eks_addon_version.default.version
}

output "latest" {
  value = data.aws_eks_addon_version.latest.version
}
```

## Argument Reference

* `addon_name` – (Required) Name of the EKS add-on. The name must match one of
  the names returned by [list-addon](https://docs.aws.amazon.com/cli/latest/reference/eks/list-addons.html).
* `kubernetes_version` – (Required) Version of the EKS Cluster. Must be between 1-100 characters in length. Must begin with an alphanumeric character, and must only contain alphanumeric characters, dashes and underscores (`^[0-9A-Za-z][A-Za-z0-9\-_]+$`).
* `latest` - (Optional) Determine if the latest or default version of the addon should be returned.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The name of the add-on
* `version` - The version of the EKS add-on.
