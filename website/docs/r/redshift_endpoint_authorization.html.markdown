---
subcategory: "Redshift"
layout: "aws"
page_title: "AWS: aws_redshift_endpoint_authorization"
description: |-
  Provides a Redshift Endpoint Authorization resource.
---

# Resource: aws_redshift_endpoint_authorization

Creates a new Amazon Redshift endpoint authorization.

## Example Usage

```terraform
resource "aws_redshift_endpoint_authorization" "example" {
  account            = "01234567910"
  cluster_identifier = aws_redshift_cluster.example.cluster_identifier
}
```

## Argument Reference

This resource supports the following arguments:

* `account` - (Required) The Amazon Web Services account ID to grant access to.
* `cluster_identifier` - (Required) The cluster identifier of the cluster to grant access to.
* `force_delete` - (Optional) Indicates whether to force the revoke action. If true, the Redshift-managed VPC endpoints associated with the endpoint authorization are also deleted. Default value is `false`.
* `vpc_ids` - (Optional) The virtual private cloud (VPC) identifiers to grant access to. If none are specified all VPCs in shared account are allowed.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `allowed_all_vpcs` - Indicates whether all VPCs in the grantee account are allowed access to the cluster.
* `id` - The identifier of the Redshift Endpoint Authorization, `account`, and `cluster_identifier` separated by a colon (`:`).
* `endpoint_count` - The number of Redshift-managed VPC endpoints created for the authorization.
* `grantee` - The Amazon Web Services account ID of the grantee of the cluster.
* `grantor` - The Amazon Web Services account ID of the cluster owner.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Redshift endpoint authorization using the `id`. For example:

```terraform
import {
  to = aws_redshift_endpoint_authorization.example
  id = "01234567910:cluster-example-id"
}
```

Using `terraform import`, import Redshift endpoint authorization using the `id`. For example:

```console
% terraform import aws_redshift_endpoint_authorization.example 01234567910:cluster-example-id
```
