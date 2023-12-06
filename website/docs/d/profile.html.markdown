---
subcategory: "Meta Data Sources"
layout: "aws"
page_title: "AWS: aws_profile"
description: |-
    Provides the profile name used to configure the provider.
---

# Data Source: aws_profile

Supplies the profile name used to configure the provider. This can be used to pass the provider's profile to other resources, e.g. an external process using a [`provisioner`](https://developer.hashicorp.com/terraform/language/resources/provisioners/local-exec).

## Example Usage
```terraform
provider "aws" {
  profile = "default"
}

data "aws_profile" "profile" {}
```

Using local-exec to provision the local `.kube/config` file after creating an `aws_eks_cluster` resource.

```terraform
provider "aws" {
  region  = "eu-central-1"
  profile = "example"
}

data "aws_profile" "current" {}

resource "aws_eks_cluster" "cluster" {
  # ...
}

resource "null_resource" "cluster-config" {
  triggers = {
    cluster      = aws_eks_cluster.cluster.endpoint
    cluster-cert = aws_eks_cluster.cluster.certificate_authority[0].data
  }
  provisioner "local-exec" {
    command = "aws eks update-kubeconfig --region ${data.aws_region.current.name} --name ${var.cluster-name} --profile ${data.aws_profile.current.name}"
  }
}
```

## Attribute Reference

This resource exports the following attributes:

* `id` - The profile name used to configure the provider.
* `name` - The profile name used to configure the provider (same as `id`).
