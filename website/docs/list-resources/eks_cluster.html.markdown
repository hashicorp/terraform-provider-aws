---
subcategory: "EKS (Elastic Kubernetes)"
layout: "aws"
page_title: "AWS: aws_eks_cluster"
description: |-
  Lists EKS (Elastic Kubernetes) Cluster resources.
---

# List Resource: aws_eks_cluster

Lists EKS (Elastic Kubernetes) Cluster resources.

## Example Usage

```terraform
list "aws_eks_cluster" "example" {
  provider = aws
}
```

## Argument Reference

This list resource supports the following arguments:

* `region` - (Optional) Region to query. Defaults to provider region.
