---
subcategory: "EKS (Elastic Kubernetes)"
layout: "aws"
page_title: "AWS: aws_eks_access_entry"
description: |-
  Access Entry Configurations for an EKS Cluster.
---

# Resource: aws_eks_access_entry

Access Entry Configurations for an EKS Cluster.

## Example Usage

```terraform
resource "aws_eks_access_entry" "example" {
  cluster_name      = aws_eks_cluster.example.name
  principal_arn     = aws_iam_role.example.arn
  kubernetes_groups = ["group-1", "group-2"]
  type              = "STANDARD"
}
```

## Argument Reference

The following arguments are required:

* `cluster_name` - (Required) Name of the EKS Cluster.
* `principal_arn` - (Required) The IAM Principal ARN which requires Authentication access to the EKS cluster.

The following arguments are optional:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `kubernetes_groups` - (Optional) List of string which can optionally specify the Kubernetes groups the user would belong to when creating an access entry.
* `tags` - (Optional) Key-value map of resource tags. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.
* `type` - (Optional) Defaults to STANDARD which provides the standard workflow. EC2_LINUX, EC2_WINDOWS, FARGATE_LINUX types disallow users to input a username or groups, and prevent associations.
* `user_name` - (Optional) Defaults to principal ARN if user is principal else defaults to assume-role/session-name is role is used.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `access_entry_arn` - Amazon Resource Name (ARN) of the Access Entry.
* `created_at` - Date and time in [RFC3339 format](https://tools.ietf.org/html/rfc3339#section-5.8) that the EKS add-on was created.
* `modified_at` - Date and time in [RFC3339 format](https://tools.ietf.org/html/rfc3339#section-5.8) that the EKS add-on was updated.
* `tags_all` - (Optional) Key-value map of resource tags, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `20m`)
* `update` - (Default `20m`)
* `delete` - (Default `40m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import EKS add-on using the `cluster_name` and `principal_arn` separated by a colon (`:`). For example:

```terraform
import {
  to = aws_eks_access_entry.my_eks_entry
  id = "my_cluster_name:my_principal_arn"
}
```

Using `terraform import`, import EKS access entry using the `cluster_name` and `principal_arn` separated by a colon (`:`). For example:

```console
% terraform import aws_eks_access_entry.my_eks_access_entry my_cluster_name:my_principal_arn
```
