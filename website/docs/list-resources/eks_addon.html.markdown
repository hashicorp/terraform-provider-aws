---
subcategory: "EKS (Elastic Kubernetes)"
layout: "aws"
page_title: "AWS: aws_eks_addon"
description: |-
  Lists EKS (Elastic Kubernetes) Addon resources for a cluster.
---

# List Resource: aws_eks_addon

Lists EKS (Elastic Kubernetes) Addon resources for a cluster.

## Example Usage

```terraform
list "aws_eks_addon" "example" {
  provider = aws

  config {
    cluster_name = aws_eks_cluster.example.name
  }
}
```

## Argument Reference

This list resource supports the following arguments:

* `cluster_name` - (Required) Name of the EKS Cluster to list add-ons for.
* `region` - (Optional) Region to query. Defaults to provider region.
