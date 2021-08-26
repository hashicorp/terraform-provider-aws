---
subcategory: "EKS"
layout: "aws"
page_title: "AWS: aws_eks_clusters"
description: |-
  Retrieve EKS Clusters list
---

# Data Source: aws_eks_cluster

Retrieve EKS Clusters list

## Example Usage

```terraform
data "aws_eks_clusters" "example" {}

data "aws_eks_cluster" "example" {
  for_each = toset(data.aws_eks_clusters.example.clusters)
  name     = each.value
}
```

## Attributes Reference

* `clusters` - The list of EKS clusters names
