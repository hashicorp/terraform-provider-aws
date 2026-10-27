---
subcategory: "EKS (Elastic Kubernetes)"
layout: "aws"
page_title: "AWS: aws_eks_access_policy_association"
description: |-
  Lists EKS Access Policy Association resources.
---

# List Resource: aws_eks_access_policy_association

Lists EKS Access Policy Association resources.

## Example Usage

```terraform
list "aws_eks_access_policy_association" "example" {
  provider = aws

  config {
    cluster_name  = aws_eks_cluster.example.name
    principal_arn = aws_eks_access_entry.example.principal_arn
  }
}
```

## Argument Reference

This list resource supports the following arguments:

* `cluster_name` - (Required) Name of the cluster to list access policy associations from.
* `principal_arn` - (Required) ARN of the IAM principal for the access entry.
* `region` - (Optional) Region to query. Defaults to provider region.
