---
subcategory: "Redshift"
layout: "aws"
page_title: "AWS: aws_redshift_partner"
description: |-
  Provides a Redshift Partner resource.
---

# Resource: aws_redshift_partner

Creates a new Amazon Redshift Partner Integration.

## Example Usage

```terraform
resource "aws_redshift_partner" "example" {
  cluster_identifier = aws_redshift_cluster.example.id
  account_id         = 01234567910
  database_name      = aws_redshift_cluster.example.database_name
  partner_name       = "example"
}
```

## Argument Reference

This resource supports the following arguments:

* `account_id` - (Required) The Amazon Web Services account ID that owns the cluster.
* `cluster_identifier` - (Required) The cluster identifier of the cluster that receives data from the partner.
* `database_name` - (Required) The name of the database that receives data from the partner.
* `partner_name` - (Required) The name of the partner that is authorized to send data.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - The identifier of the Redshift partner, `account_id`, `cluster_identifier`, `database_name`, `partner_name` separated by a colon (`:`).
* `status` - (Optional) The partner integration status.
* `status_message` - (Optional) The status message provided by the partner.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Redshift usage limits using the `id`. For example:

```terraform
import {
  to = aws_redshift_partner.example
  id = "01234567910:cluster-example-id:example:example"
}
```

Using `terraform import`, import Redshift usage limits using the `id`. For example:

```console
% terraform import aws_redshift_partner.example 01234567910:cluster-example-id:example:example
```
