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

This resource supports the following arguments:

* `cluster_identifier` - (Required) The name of the Redshift Cluster IAM Roles.
* `iam_role_arns` - (Optional) A list of IAM Role ARNs to associate with the cluster. A Maximum of 10 can be associated to the cluster at any time.
* `default_iam_role_arn` - (Optional) The Amazon Resource Name (ARN) for the IAM role that was set as default for the cluster when the cluster was created.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - The Redshift Cluster ID.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Redshift Cluster IAM Roless using the `cluster_identifier`. For example:

```terraform
import {
  to = aws_redshift_cluster_iam_roles.examplegroup1
  id = "example"
}
```

Using `terraform import`, import Redshift Cluster IAM Roless using the `cluster_identifier`. For example:

```console
% terraform import aws_redshift_cluster_iam_roles.examplegroup1 example
```
