---
subcategory: "EKS (Elastic Kubernetes)"
layout: "aws"
page_title: "AWS: aws_eks_access_policies"
description: |-
  Terraform data source for managing AWS EKS (Elastic Kubernetes) Access Policies.
---

# Data Source: aws_eks_access_policies

Terraform data source for managing AWS EKS (Elastic Kubernetes) Access Policies.

## Example Usage

### Basic Usage

```terraform
data "aws_eks_access_policies" "example" {}

output "eks_access_policies" {
  value = data.aws_eks_access_policies.example.access_policies
}

output "eks_networking_policy" {
  value = [for ap in data.aws_eks_access_policies.example.access_policies : ap if ap.name == "AmazonEKSNetworkingPolicy"]
}

output "eks_access_policy_names" {
  value = [for ap in data.aws_eks_access_policies.example.access_policies : ap.name]
}
```

## Argument Reference

The following arguments are optional:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `access_policies` - List of available access policies.
    * `arn` - ARN of the access policy.
    * `name` - Name of the access policy.
