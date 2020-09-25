---
subcategory: "VPC"
layout: "aws"
page_title: "AWS: aws_main_route_table_association"
description: |-
  Provides a resource for managing the main routing table of a VPC.
---

# Resource: aws_main_route_table_association

Provides a resource for managing the main routing table of a VPC.

It is recommended you **do not** use both `aws_main_route_table_association` to
manage the default route table **and** use the `aws_default_route_table`,
due to possible conflict in routes.

For more information about Route Tables, see the AWS Documentation on
[Route Tables][aws-route-tables].

For more information about managing normal Route Tables in Terraform, see our
documentation on [aws_route_table][tf-route-tables].

## Example Usage

```hcl
resource "aws_main_route_table_association" "a" {
  vpc_id         = aws_vpc.foo.id
  route_table_id = aws_route_table.bar.id
}
```

## Argument Reference

The following arguments are supported:

* `vpc_id` - (Required) The ID of the VPC whose main route table should be set
* `route_table_id` - (Required) The ID of the Route Table to set as the new
  main route table for the target VPC

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The ID of the Route Table Association
* `original_route_table_id` - Used internally, see __Notes__ below

## Notes

On VPC creation, the AWS API always creates an initial Main Route Table. This
resource records the ID of that Route Table under `original_route_table_id`.
The "Delete" action for a `main_route_table_association` consists of resetting
this original table as the Main Route Table for the VPC. You'll see this
additional Route Table in the AWS console; it must remain intact in order for
the `main_route_table_association` delete to work properly.

[aws-route-tables]: http://docs.aws.amazon.com/AmazonVPC/latest/UserGuide/VPC_Route_Tables.html#Route_Replacing_Main_Table
[tf-route-tables]: /docs/providers/aws/r/route_table.html