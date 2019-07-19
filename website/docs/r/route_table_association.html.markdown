---
layout: "aws"
page_title: "AWS: aws_route_table_association"
sidebar_current: "docs-aws-resource-route-table-association"
description: |-
  Provides a resource to create an association between a subnet and routing table.
---

# Resource: aws_route_table_association

Provides a resource to create an association between a subnet and routing table.

## Example Usage

```hcl
resource "aws_route_table_association" "a" {
  subnet_id      = "${aws_subnet.foo.id}"
  route_table_id = "${aws_route_table.bar.id}"
}
```

## Argument Reference

The following arguments are supported:

* `subnet_id` - (Required) The subnet ID to create an association.
* `route_table_id` - (Required) The ID of the routing table to associate with.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The ID of the association

## Import

~> **NOTE:** Attempting to associate a route table with a subnet, where either
is already associated, will result in an error (e.g.,
`Resource.AlreadyAssociated: the specified association for route table
rtb-4176657279 conflicts with an existing association`) unless you first
import the original association.

Route table associations can be imported using the subnet and route table IDs.
For example, use this command:

```
$ terraform import aws_route_table_association.assoc subnet-6777656e646f6c796e/rtb-656c65616e6f72
```
