---
subcategory: "EKS (Elastic Kubernetes)"
layout: "aws"
page_title: "AWS: aws_eks_access_entry"
description: |-
  Lists EKS Access Entry resources.
---

# List Resource: aws_eks_access_entry

Lists EKS Access Entry resources.

## Example Usage

```terraform
list "aws_eks_access_entry" "example" {
  provider = aws

  config {
    cluster_name = aws_eks_cluster.example.name
  }
}
```

## Argument Reference

This list resource supports the following arguments:

* `associated_policy_arn` - (Optional) Only access entries associated to this access policy are returned.
* `cluster_name` - (Required) Name of the cluster to list access entries from.
* `region` - (Optional) Region to query. Defaults to provider region.
