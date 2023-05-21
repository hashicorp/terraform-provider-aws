---
subcategory: "Redshift"
layout: "aws"
page_title: "AWS: aws_redshift_cluster_iam_roles"
description: |-
  Provides a Redshift Cluster IAM Roles resource.
---

# Resource: aws_redshift_cluster_iam_roles

Provides a Redshift Cluster IAM Roles resource.

~> **NOTE:** A Redshift cluster's default IAM role can be managed both by this resource's `default_iam_role_arn` argument and the [`aws_redshift_cluster`](redshift_cluster.html) resource's `default_iam_role_arn` argument. Do not configure different values for both arguments. Doing so will cause a conflict of default IAM roles.

## Example Usage

```terraform
resource "aws_redshift_cluster_iam_roles" "example" {
  cluster_identifier = aws_redshift_cluster.example.cluster_identifier
  iam_role_arns      = [aws_iam_role.example.arn]
}
```

## Argument Reference

The following arguments are supported:

* `cluster_identifier` - (Required) The name of the Redshift Cluster IAM Roles.
* `iam_role_arns` - (Optional) A list of IAM Role ARNs to associate with the cluster. A Maximum of 10 can be associated to the cluster at any time.
* `default_iam_role_arn` - (Optional) The Amazon Resource Name (ARN) for the IAM role that was set as default for the cluster when the cluster was created.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The Redshift Cluster ID.

## Import

Redshift Cluster IAM Roless can be imported using the `cluster_identifier`, e.g.,

```
$ terraform import aws_redshift_cluster_iam_roles.examplegroup1 example
```
