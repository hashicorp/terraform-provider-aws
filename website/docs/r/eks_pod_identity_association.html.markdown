---
subcategory: "EKS (Elastic Kubernetes)"
layout: "aws"
page_title: "AWS: aws_eks_pod_identity_association"
description: |-
  Terraform resource for managing an AWS EKS (Elastic Kubernetes) Pod Identity Association.
---

# Resource: aws_eks_pod_identity_association

Terraform resource for managing an AWS EKS (Elastic Kubernetes) Pod Identity Association.

Creates an EKS Pod Identity association between a service account in an Amazon EKS cluster and an IAM role with EKS Pod Identity. Use EKS Pod Identity to give temporary IAM credentials to pods and the credentials are rotated automatically.

Amazon EKS Pod Identity associations provide the ability to manage credentials for your applications, similar to the way that EC2 instance profiles provide credentials to Amazon EC2 instances.

If a pod uses a service account that has an association, Amazon EKS sets environment variables in the containers of the pod. The environment variables configure the Amazon Web Services SDKs, including the Command Line Interface, to use the EKS Pod Identity credentials.

Pod Identity is a simpler method than IAM roles for service accounts, as this method doesn’t use OIDC identity providers. Additionally, you can configure a role for Pod Identity once, and reuse it across clusters.

## Example Usage

### Basic Usage

```terraform
data "aws_iam_policy_document" "assume_role" {
  statement {
    effect = "Allow"

    principals {
      type        = "Service"
      identifiers = ["pods.eks.amazonaws.com"]
    }

    actions = [
      "sts:AssumeRole",
      "sts:TagSession"
    ]
  }
}

resource "aws_iam_role" "example" {
  name               = "eks-pod-identity-example"
  assume_role_policy = data.aws_iam_policy_document.assume_role.json
}

resource "aws_iam_role_policy_attachment" "example_s3" {
  policy_arn = "arn:aws:iam::aws:policy/AmazonS3ReadOnlyAccess"
  role       = aws_iam_role.example.name
}

resource "aws_eks_pod_identity_association" "example" {
  cluster_name    = aws_eks_cluster.example.name
  namespace       = "example"
  service_account = "example-sa"
  role_arn        = aws_iam_role.example.arn
}
```

## Argument Reference

The following arguments are required:

* `cluster_name` - (Required) The name of the cluster to create the association in.
* `namespace` - (Required) The name of the Kubernetes namespace inside the cluster to create the association in. The service account and the pods that use the service account must be in this namespace.
* `role_arn` - (Required) The Amazon Resource Name (ARN) of the IAM role to associate with the service account. The EKS Pod Identity agent manages credentials to assume this role for applications in the containers in the pods that use this service account.
* `service_account` - (Required) The name of the Kubernetes service account inside the cluster to associate the IAM credentials with.

The following arguments are optional:

* `tags` - (Optional) Key-value map of resource tags. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `association_arn` - The Amazon Resource Name (ARN) of the association.
* `association_id` - The ID of the association.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import EKS (Elastic Kubernetes) Pod Identity Association using the `cluster_name` and `association_id` separated by a comma (`,`). For example:

```terraform
import {
  to = aws_eks_pod_identity_association.example
  id = "example,a-12345678"
}
```

Using `terraform import`, import EKS (Elastic Kubernetes) Pod Identity Association using the `cluster_name` and `association_id` separated by a comma (`,`). For example:

```console
% terraform import aws_eks_pod_identity_association.example example,a-12345678
```
