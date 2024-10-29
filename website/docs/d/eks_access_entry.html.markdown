---
subcategory: "EKS (Elastic Kubernetes)"
layout: "aws"
page_title: "AWS: aws_eks_access_entry"
description: |-
  Access Entry Configurations for an EKS Cluster.
---

# Data Source: aws_eks_access_entry

Access Entry Configurations for an EKS Cluster.

## Example Usage

```terraform
data "aws_eks_access_entry" "example" {
  cluster_name  = aws_eks_cluster.example.name
  principal_arn = aws_iam_role.example.arn
}

output "eks_access_entry_outputs" {
  value = aws_eks_access_entry.example
}
```

## Argument Reference

* `cluster_name` – (Required) Name of the EKS Cluster.
* `principal_arn` – (Required) The IAM Principal ARN which requires Authentication access to the EKS cluster.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `access_entry_arn` - Amazon Resource Name (ARN) of the Access Entry.
* `created_at` - Date and time in [RFC3339 format](https://tools.ietf.org/html/rfc3339#section-5.8) that the EKS add-on was created.
* `kubernetes_groups` – List of string which can optionally specify the Kubernetes groups the user would belong to when creating an access entry.
* `modified_at` - Date and time in [RFC3339 format](https://tools.ietf.org/html/rfc3339#section-5.8) that the EKS add-on was updated.
* `user_name` - Defaults to principal ARN if user is principal else defaults to assume-role/session-name is role is used.
* `type` - Defaults to STANDARD which provides the standard workflow. EC2_LINUX, EC2_WINDOWS, FARGATE_LINUX types disallow users to input a username or groups, and prevent associations.
* `tags_all` - (Optional) Key-value map of resource tags, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).
