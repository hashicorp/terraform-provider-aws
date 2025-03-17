---
subcategory: "EKS (Elastic Kubernetes)"
layout: "aws"
page_title: "AWS: aws_eks_cluster_auth"
description: |-
  Retrieve an authentication token to communicate with an EKS cluster.
---

# Ephemeral: aws_eks_cluster_auth

Retrieve an authentication token to communicate with an EKS cluster.

~> **NOTE:** Ephemeral resources are a new feature and may evolve as we continue to explore their most effective uses. [Learn more](https://developer.hashicorp.com/terraform/language/v1.10.x/resources/ephemeral).

## Example Usage

```terraform
ephemeral "aws_eks_cluster_auth" "example" {
  name = data.aws_eks_cluster.example.id
}

provider "kubernetes" {
  host                   = data.aws_eks_cluster.example.endpoint
  cluster_ca_certificate = base64decode(data.aws_eks_cluster.example.certificate_authority[0].data)
  token                  = ephemeral.aws_eks_cluster_auth.example.token
}

provider "helm" {
  kubernetes {
    host                   = data.aws_eks_cluster.example.endpoint
    cluster_ca_certificate = base64decode(data.aws_eks_cluster.example.certificate_authority[0].data)
    token                  = ephemeral.aws_eks_cluster_auth.example.token
  }
}
```

## Argument Reference

* `name` - (Required) Name of the EKS cluster.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `token` - Token to use to authenticate with the cluster.
