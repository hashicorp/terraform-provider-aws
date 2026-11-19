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
  display_name    = "example"
  odb_network_id  = "my-odb-network-id"
  peer_network_id = "my-vpc-id"
  tags = {
    "env" = "dev"
  }
}
```

## Argument Reference

The following arguments are required:

* `display_name` - (Required) Display name of the ODB network peering connection. Changing this will force Terraform to create a new resource.
* `peer_network_id` - (Required) ID of the VPC or ODB network to peer with. Changing this replaces the peering connection.

The following arguments are optional:

* `odb_network_arn` - (Optional) ARN of the ODB network that initiates the peering connection. Exactly one of `odb_network_id` and `odb_network_arn` must be configured. Changing this replaces the peering connection.
* `odb_network_id` - (Optional) ID or ARN of the ODB network that initiates the peering connection. Changing to a different ODB network replaces the peering connection. Switching between the ID and ARN of the same network updates Terraform state without changing the peering connection.
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

AWS does not return whether the peering connection was created with an ODB network ID or ARN. Import sets `odb_network_id` to the ID. If the configuration uses the ARN in `odb_network_id`, Terraform performs a one-time in-place state update to retain the configured ARN. The same update can occur when upgrading a peering connection whose state was written by an older provider version. Subsequent plans have no change.

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import an ODB network peering connection using its `id`. For example:

```terraform
import {
  to = aws_odb_network_peering_connection.example
  id = "odbpcx_abcdefgh12"
}
```

Using `terraform import`, import odb network peering using the `id`. For example:

```console
% terraform import aws_odb_network_peering_connection.example odbpcx_abcdefgh12
```
