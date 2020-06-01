---
subcategory: "EKS"
layout: "aws"
page_title: "AWS: aws_eks_node_group_names"
description: |-
  Provides a set of node group names for a EKS Cluster
---

# Data Source: aws_eks_node_group_names

Retrieve information about an EKS Node Group names.

## Example Usage

```hcl
data "aws_eks_node_group_names" "example" {
  cluster_name = "example"
}

output "node_groups" {
  value = "${data.aws_eks_node_group_names.example.names}"
}
```

```hcl
data "aws_eks_node_group_names" "example" {
  cluster_name = "example"
}

data "aws_eks_node_group" "example" {
  for_each = data.aws_eks_node_group_names.example.names

  cluster_name    = "example"
  node_group_name = each.value
}
```


## Argument Reference

* `cluster_name` - (Required) The name of the cluster.

## Attributes Reference

* `names` - A set of all node group names in an EKS Cluster.
