---
subcategory: "EKS (Elastic Kubernetes)"
layout: "aws"
page_title: "AWS: aws_eks_addon"
description: |-
  Lists EKS Add On resources.
---

# List Resource: aws_eks_addon

Lists EKS (Elastic Kubernetes) Add-On resources.

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

* `cluster_name` - (Required) Name of the cluster to list add-ons from.
* `region` - (Optional) Region to query. Defaults to provider region.
