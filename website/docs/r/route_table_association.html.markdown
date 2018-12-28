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

## Example Replacement Usage

If the subnet already has an associated route table, normally an error will be thrown if attempting to associate another route table. However, using `force_replace`, no error will be thrown and the subnet will become associated with the new route table instead.

```hcl
resource "aws_route_table_association" "a" {
  subnet_id      = "${aws_subnet.foo.id}"
  route_table_id = "${aws_route_table.bar.id}"
  force_replace  = true
}
```

## Argument Reference

The following arguments are supported:

* `subnet_id` - (Required) The subnet ID to create an association.
* `route_table_id` - (Required) The ID of the routing table to associate with.
* `force_replace` - (Optional) Boolean indicating whether to replace an existing association or not.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The ID of the association

