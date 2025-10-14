---
subcategory: "VPC (Virtual Private Cloud)"
layout: "aws"
page_title: "AWS: aws_vpc_route_server"
description: |-
  Terraform resource for managing a VPC (Virtual Private Cloud) Route Server.
---
# Resource: aws_vpc_route_server

  Provides a resource for managing a VPC (Virtual Private Cloud) Route Server.

## Example Usage

### Basic Usage

```terraform
resource "aws_vpc_route_server" "test" {
  amazon_side_asn = 65534

  tags = {
    Name = "Test"
  }
}
```

### Persist Route and SNS Notification

```terraform
resource "aws_vpc_route_server" "test" {
  amazon_side_asn           = 65534
  persist_routes            = "enable"
  persist_routes_duration   = 2
  sns_notifications_enabled = true

  tags = {
    Name = "Main Route Server"
  }
}
```

## Argument Reference

The following arguments are required:

* `amazon_side_asn` - (Required) The Border Gateway Protocol (BGP) Autonomous System Number (ASN) for the appliance. Valid values are from 1 to 4294967295.

The following arguments are optional:

* `persist_routes` - (Optional) Indicates whether routes should be persisted after all BGP sessions are terminated. Valid values are `enable`, `disable`, `reset`
* `persist_routes_duration` - (Optional) The number of minutes a route server will wait after BGP is re-established to unpersist the routes in the FIB and RIB. Value must be in the range of 1-5. Required if `persist_routes` is enabled.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `sns_notifications_enabled` - (Optional) Indicates whether SNS notifications should be enabled for route server events. Enabling SNS notifications persists BGP status changes to an SNS topic provisioned by AWS`.
* `tags` - (Optional) A map of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - The ARN of the route server.
* `route_server_id` - The unique identifier of the route server.
* `sns_topic_arn` - The ARN of the SNS topic where notifications are published.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `30m`)
* `update` - (Default `30m`)
* `delete` - (Default `30m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import VPC (Virtual Private Cloud) Route Server using the `route_server_id`. For example:

```terraform
import {
  to = aws_vpc_route_server.example
  id = "rs-12345678"
}
```

Using `terraform import`, import VPC (Virtual Private Cloud) Route Server using the `route_server_id`. For example:

```console
% terraform import aws_vpc_route_server.example rs-12345678
```
