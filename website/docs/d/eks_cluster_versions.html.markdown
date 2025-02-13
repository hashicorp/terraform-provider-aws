---
subcategory: "EKS (Elastic Kubernetes)"
layout: "aws"
page_title: "AWS: aws_eks_cluster_versions"
description: |-
  Terraform data source for managing an AWS EKS (Elastic Kubernetes) Cluster Versions.
---

# Data Source: aws_eks_cluster_versions

Terraform data source for managing an AWS EKS (Elastic Kubernetes) Cluster Versions.

## Example Usage

### Basic Usage

```terraform
data "aws_eks_cluster_versions" "example" {}

data "aws_eks_cluster_versions" "example" {
  cluster_type = "eks"
}
```

## Argument Reference

The following arguments are optional:

* `cluster_type` - (Optional) The type of clusters to filter by. Currently only `eks` is supported.
* `default_only` - (Optional) Whether to show only the default versions of Kubernetes supported by EKS. Default is `false`.
* `cluster_versions` - (Optional) A list of Kubernetes versions that you can use to check if EKS supports it.
* `include_all` - (Optional) Whether to include all kubernetes versions in the response. Default is `false`.
* `status` - (Optional) The status of the EKS cluster versions to list. Can be `STANDARD_SUPPORT` or `UNSUPPORTED` or `EXTENDED_SUPPORT`.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `cluster_version` - The Kubernetes version supported by EKS.
* `cluster_type` - The type of cluster that the version belongs to. Currently only `eks` is supported.
* `default_platform_version` - The default eks platform version for the cluster version.
* `default_version` - The default Kubernetes version for the cluster version.
* `status` - The status of the EKS cluster version. Can be `STANDARD_SUPPORT` or `UNSUPPORTED` or `EXTENDED_SUPPORT`.
* `end_of_extended_support_date` - The end of extended support date for the cluster version.
* `end_of_standard_support_date` - The end of standard support date for the cluster version.
* `kubernetes_patch_version` - The Kubernetes patch version for the cluster version.
* `release_date` - The release date of the cluster version.
