---
subcategory: "EKS (Elastic Kubernetes)"
layout: "aws"
page_title: "AWS: aws_eks_node_group"
description: |-
  Lists EKS Node Group resources.
---

# List Resource: aws_eks_node_group

Lists EKS (Elastic Kubernetes) Node Group resources.

## Example Usage

```terraform
list "aws_eks_node_group" "example" {
  provider = aws

  config {
    cluster_name = aws_eks_cluster.example.name
  }
}
```

## Argument Reference

This list resource supports the following arguments:

* `cluster_name` - (Required) Name of the cluster to list node groups from.
* `region` - (Optional) Region to query. Defaults to provider region.
