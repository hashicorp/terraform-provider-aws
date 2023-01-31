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

The following arguments are supported:

* `account_id` - (Required) The Amazon Web Services account ID that owns the cluster.
* `cluster_identifier` - (Required) The cluster identifier of the cluster that receives data from the partner.
* `database_name` - (Required) The name of the database that receives data from the partner.
* `partner_name` - (Required) The name of the partner that is authorized to send data.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The identifier of the Redshift partner, `account_id`, `cluster_identifier`, `database_name`, `partner_name` separated by a colon (`:`).
* `status` - (Optional) The partner integration status.
* `status_message` - (Optional) The status message provided by the partner.

## Import

Redshift usage limits can be imported using the `id`, e.g.,

```
$ terraform import aws_redshift_partner.example 01234567910:cluster-example-id:example:example
```
