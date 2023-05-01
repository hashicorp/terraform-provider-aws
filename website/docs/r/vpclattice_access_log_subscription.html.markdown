---
subcategory: "VPC Lattice"
layout: "aws"
page_title: "AWS: aws_vpclattice_access_log_subscription"
description: |-
  Terraform resource for managing an AWS VPC Lattice Service Network or Services Access log subscription.
---

# Resource: aws_vpclattice_access_log_subscription

Terraform resource for managing an AWS VPC Lattice Service Network or Service Access log subscription.

## Example Usage

### Basic Usage

```terraform
resource "aws_vpclattice_access_log_subscription" "example" {
  resource_identifier = aws_vpclattice_service_network.example.id
  destination_arn     = aws_s3.bucket.arn
}
```

## Argument Reference

The following arguments are required:

* `destination_arn` - (Required) Amazon Resource Name (ARN) of the log destination.
* `resource_identifier` - (Required) The ID or Amazon Resource Identifier (ARN) of the service network or service. You must use the ARN if the resources specified in the operation are in different accounts.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - ID of the access log subscription.
* `arn` - Amazon Resource Name (ARN) of the access log subscription.
* `resource_identifier` - ID of the service network or service.
* `resource_arn` - Amazon Resource Name (ARN) of the service network or service.
* `destination_arn` - Amazon Resource Name (ARN) of the log destination.

## Import

VPC Lattice Access Log Subscription can be imported using the access log subscription ID, e.g.,

```
$ terraform import aws_vpclattice_access_log_subscription.example rft-8012925589
```
