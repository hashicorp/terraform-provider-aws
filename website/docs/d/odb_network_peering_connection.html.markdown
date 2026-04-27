---
subcategory: "Oracle Database@AWS"
layout: "AWS: aws_odb_network_peering_connection"
page_title: "AWS: aws_odb_network_peering_connection"
description: |-
  Terraform data source for managing oracle database network peering resource in AWS.
---

# Data Source: aws_odb_network_peering_connection

Terraform data source for managing oracle database network peering resource in AWS.

You can find out more about Oracle Database@AWS from [User Guide](https://docs.aws.amazon.com/odb/latest/UserGuide/what-is-odb.html).

## Example Usage

### Basic Usage

```terraform
data "aws_odb_network_peering_connection" "example" {
  id = "example"
}
```

## Argument Reference

The following arguments are required:

* `id` - (Required) The unique identifier of the ODB network peering connection.

The following arguments are optional:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `display_name` - Display name of the ODB network peering connection.
* `status` - Status of the ODB network peering connection.
* `status_reason` - Status of the ODB network peering connection.
* `odb_network_arn` - ARN of the ODB network peering connection.
* `arn` - The Amazon Resource Name (ARN) for the  ODB network peering connection.
* `peer_network_arn` - ARN of the peer network peering connection.
* `odb_peering_connection_type` - Type of the ODB peering connection.
* `peer_network_cidrs` - Set of peer network cidrs.
* `created_at` - Created time of the ODB network peering connection.
* `percent_progress` - Progress of the ODB network peering connection.
* `tags` - Tags applied to the resource.  
