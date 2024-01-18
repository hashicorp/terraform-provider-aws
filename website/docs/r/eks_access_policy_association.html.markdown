---
subcategory: "EKS (Elastic Kubernetes)"
layout: "aws"
page_title: "AWS: aws_eks_access_policy_associattion"
description: |-
  Access Entry Policy Association for an EKS Cluster.
---

# Resource: aws_eks_access_policy_association

Access Entry Policy Association for an EKS Cluster.

## Example Usage

```terraform
resource "aws_eks_access_policy_association" "example" {
  cluster_name  = aws_eks_cluster.example.name
  policy_arn    = "arn:aws:eks::aws:cluster-access-policy/AmazonEKSViewPolicy"
  principal_arn = aws_iam_user.example.arn

  access_scope = {
    type       = "namespace"
    namespaces = ["example-namespace"]
  }
}
```

## Argument Reference

The following arguments are required:

* `cluster_name` – (Required) Name of the EKS Cluster. Must be between 1-100 characters in length. Must begin with an alphanumeric character, and must only contain alphanumeric characters, dashes and underscores (`^[0-9A-Za-z][A-Za-z0-9\-_]+$`).
* `policy_arn` – (Required) The ARN of the access policy that you're associating.
* `principal_arn` – (Required) The IAM Princial ARN which requires Authentication access to the EKS cluster.
* `access_scope` – (Required) The configuration block to determine the scope of the access.
    * `type` - (Required) Valid values are `namespace` or `cluster`.
    * `namespaces` - (Optional) The namespaces to which the access scope applies when type is namespace.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `associated_access_policy` - Configuraion block which consists of
    * `associated_at` - Date and time in [RFC3339 format](https://tools.ietf.org/html/rfc3339#section-5.8) that the policy was associated.
    * `modified_at` - Date and time in [RFC3339 format](https://tools.ietf.org/html/rfc3339#section-5.8) that the policy was updated.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `20m`)
* `update` - (Default `20m`)
* `delete` - (Default `40m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import EKS add-on using the `cluster_name`, `policy_arn` and `principal_arn` separated by a colon (`:`). For example:

```terraform
import {
  to = aws_eks_access_policy_association.my_eks_entry
  id = "my_cluster_name:my_policy_arn:my_principal_arn"
}
```

Using `terraform import`, import EKS access entry using the `cluster_name` `policy_arn` and `principal_arn` separated by a colon (`:`). For example:

```console
% terraform import aws_eks_access_policy_association.my_eks_access_entry my_cluster_name:my_policy_arn:my_principal_arn
```
