---
subcategory: "Direct Connect"
layout: "aws"
page_title: "AWS: aws_dx_lag"
description: |-
  Provides a Direct Connect LAG.
---

# Resource: aws_dx_lag

Provides a Direct Connect LAG. Connections can be added to the LAG via the [`aws_dx_connection`](/docs/providers/aws/r/dx_connection.html) and [`aws_dx_connection_association`](/docs/providers/aws/r/dx_connection_association.html) resources.

~> *NOTE:* When creating a LAG, if no existing connection is specified, Direct Connect will create a connection and Terraform will remove this unmanaged connection during resource creation.

## Example Usage

```terraform
resource "aws_dx_lag" "hoge" {
  name                  = "tf-dx-lag"
  connections_bandwidth = "1Gbps"
  location              = "EqDC2"
  force_destroy         = true
}
```

## Argument Reference

This resource supports the following arguments:

* `name` - (Required) The name of the LAG.
* `connections_bandwidth` - (Required) The bandwidth of the individual dedicated connections bundled by the LAG. Valid values: 1Gbps, 10Gbps, 100Gbps, and 400Gbps. Case sensitive. Refer to the AWS Direct Connection supported bandwidths for [Dedicated Connections](https://docs.aws.amazon.com/directconnect/latest/UserGuide/dedicated_connection.html).
* `location` - (Required) The AWS Direct Connect location in which the LAG should be allocated. See [DescribeLocations](https://docs.aws.amazon.com/directconnect/latest/APIReference/API_DescribeLocations.html) for the list of AWS Direct Connect locations. Use `locationCode`.
* `connection_id` - (Optional) The ID of an existing dedicated connection to migrate to the LAG.
* `force_destroy` - (Optional, Default:false) A boolean that indicates all connections associated with the LAG should be deleted so that the LAG can be destroyed without error. These objects are *not* recoverable.
* `provider_name` - (Optional) The name of the service provider associated with the LAG.
* `tags` - (Optional) A map of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - The ARN of the LAG.
* `has_logical_redundancy` - Indicates whether the LAG supports a secondary BGP peer in the same address family (IPv4/IPv6).
* `id` - The ID of the LAG.
* `jumbo_frame_capable` -Indicates whether jumbo frames (9001 MTU) are supported.
* `owner_account_id` - The ID of the AWS account that owns the LAG.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Direct Connect LAGs using the LAG `id`. For example:

```terraform
import {
  to = aws_dx_lag.test_lag
  id = "dxlag-fgnsp5rq"
}
```

Using `terraform import`, import Direct Connect LAGs using the LAG `id`. For example:

```console
% terraform import aws_dx_lag.test_lag dxlag-fgnsp5rq
```
