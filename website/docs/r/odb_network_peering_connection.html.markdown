---
subcategory: "Oracle Database@AWS"
layout: "AWS: aws_odb_network_peering_connection"
page_title: "AWS: aws_odb_network_peering_connection"
description: |-
  Terraform  resource for managing oracle database network peering resource in AWS.
---

# Resource: aws_odb_network_peering_connection

Terraform  resource for managing oracle database network peering resource in AWS. If underlying odb network is shared, ARN must be used while creating network peering.

You can find out more about Oracle Database@AWS from [User Guide](https://docs.aws.amazon.com/odb/latest/UserGuide/what-is-odb.html).

## Example Usage

### Basic Usage

```terraform
resource "aws_odb_network_peering_connection" "example" {
  display_name                = "example"
  odb_network_id              = "odbnet_3l9st3litg"
  peer_network_id             = "vpc-1234567890abcdef0"
  peer_network_route_table_id = "rtb-1234567890abcdef0"
  tags = {
    "env" = "dev"
  }
}
```

For a VPC and ODB network created in the same Terraform configuration, see the [complete example](https://github.com/hashicorp/terraform-provider-aws/tree/main/examples/odb-network-peering-route-table). The peering resource can use `aws_vpc.example.main_route_table_id` as `peer_network_route_table_id`.

AWS creates a route to the ODB network in the selected VPC route table while establishing the peering. Omit `peer_network_route_table_id` if you manage VPC routes separately. The [CreateOdbPeeringConnection API](https://docs.aws.amazon.com/odb/latest/APIReference/API_CreateOdbPeeringConnection.html) accepts exactly one route table ID when this input is supplied. To route additional tables, manage their routes separately with [`aws_route`](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/route) or [`aws_route_table`](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/route_table); do not also manage the automatically created route in the selected table.

The IAM principal needs `odb:CreateOdbPeeringConnection`, `odb:GetOdbPeeringConnection`, `odb:UpdateOdbPeeringConnection`, and `odb:DeleteOdbPeeringConnection`, plus `odb:ListTagsForResource` for resource reads and `odb:TagResource` and `odb:UntagResource` when managing tags. For automatic route management, the [AWS example policy](https://docs.aws.amazon.com/odb/latest/UserGuide/security_iam_id-based-policy-examples.html) also includes `ec2:DescribeRouteTables`, `ec2:CreateRoute`, `ec2:DeleteRoute`, `ec2:CreateOdbNetworkPeering`, and `ec2:DeleteOdbNetworkPeering`. Initial ODB service setup may require `iam:CreateServiceLinkedRole`.

## Argument Reference

The following arguments are required:

* `display_name` - (Required) Display name of the ODB network peering connection. Changing this will force Terraform to create a new resource.
* `peer_network_id` - (Required) ID of the VPC or ODB network to peer with. Changing this replaces the peering connection.

The following arguments are optional:

* `odb_network_arn` - (Optional) ARN of the ODB network that initiates the peering connection. Changing this will force Terraform to create a new resource. Either odb_network_id or odb_network_arn should be used.
* `odb_network_id` - (Optional) ID of the ODB network that initiates the peering connection. A sample ID is `odbnet_3l9st3litg`. Exactly one of `odb_network_id` and `odb_network_arn` must be configured. Changing this replaces the peering connection.
* `peer_network_route_table_id` - (Optional) VPC route table ID where AWS creates a route to the ODB network during peering creation. Must match `rtb-[a-z0-9]{8,17}`. Only one route table is supported. Adding, removing, or changing this value replaces the peering connection because the AWS Update API cannot change it.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `tags` - (Optional) Map of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the ODB network peering connection.
* `created_at` - Created time of the ODB network peering connection.
* `id` - Unique identifier of odb network peering connection.
* `odb_peering_connection_type` - Type of the ODB peering connection.
* `peer_network_arn` - ARN of the peer network peering connection.
* `peer_network_cidrs` - Set of peer network cidrs. Add remove is only supported during update operation. During create this attribute is compute only.
* `percent_progress` - Progress of the ODB network peering connection.
* `status` - Status of the ODB network peering connection.
* `status_reason` - Reason for the current status of the ODB peering connection.
* `tags_all` - Map of tags assigned to the resource, including inherited tags.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `24h`)
* `update` - (Default `24h`)
* `delete` - (Default `24h`)

## Import

Import identifies the peering connection by ID. The [AWS Get API](https://docs.aws.amazon.com/odb/latest/APIReference/API_GetOdbPeeringConnection.html) does not return the route table selected at creation, so an imported peering has no `peer_network_route_table_id` in Terraform state. Setting it afterward plans replacement, even if that table already contains the route. Terraform retains the configured ID for managed peerings but cannot detect route removal from the route table during peering refresh. When a peering created with this argument is deleted, AWS removes the automatically created route.

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import an ODB network peering connection using the `id`. For example:

```terraform
import {
  to = aws_odb_network_peering_connection.example
  id = "odbpcx_abcdefghij"
}
```

Using `terraform import`, import odb network peering using the `id`. For example:

```console
% terraform import aws_odb_network_peering_connection.example odbpcx_abcdefghij
```
