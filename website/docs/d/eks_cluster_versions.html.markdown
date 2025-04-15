---
subcategory: "EKS (Elastic Kubernetes)"
layout: "aws"
page_title: "AWS: aws_eks_cluster_versions"
description: |-
  Terraform data source for managing AWS EKS (Elastic Kubernetes) Cluster Versions.
---

# Data Source: aws_eks_cluster_versions

Terraform data source for managing AWS EKS (Elastic Kubernetes) Cluster Versions.

## Example Usage

### Basic Usage

```terraform
data "aws_eks_cluster_versions" "example" {}
```

### Filter by Cluster Type

```terraform
data "aws_eks_cluster_versions" "example" {
  cluster_type = "eks"
}
```

### Filter by Version Status

```terraform
data "aws_eks_cluster_versions" "example" {
  version_status = "STANDARD_SUPPORT"
}
```

## Argument Reference

The following arguments are optional:

* `cluster_type` - (Optional) Type of clusters to filter by.
Currently, the only valid value is `eks`.
* `cluster_versions` - (Optional) A list of Kubernetes versions that you can use to check if EKS supports it.
* `default_only` - (Optional) Whether to show only the default versions of Kubernetes supported by EKS.
* `include_all` - (Optional) Whether to include all kubernetes versions in the response.
* `version_status` - (Optional) Status of the EKS cluster versions to list.
Valid values are `STANDARD_SUPPORT` or `UNSUPPORTED` or `EXTENDED_SUPPORT`.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `cluster_type` - Type of cluster that the version belongs to.
* `cluster_version` - Kubernetes version supported by EKS.
* `default_platform_version` - Default eks platform version for the cluster version.
* `default_version` - Default Kubernetes version for the cluster version.
* `end_of_extended_support_date` - End of extended support date for the cluster version.
* `end_of_standard_support_date` - End of standard support date for the cluster version.
* `kubernetes_patch_version` - Kubernetes patch version for the cluster version.
* `release_date` - Release date of the cluster version.
* `version_status` - Status of the EKS cluster version.
