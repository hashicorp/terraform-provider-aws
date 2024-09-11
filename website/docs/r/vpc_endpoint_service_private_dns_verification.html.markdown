---
subcategory: "VPC (Virtual Private Cloud)"
layout: "aws"
page_title: "AWS: aws_vpc_endpoint_service_private_dns_verification"
description: |-
  Terraform resource for managing an AWS VPC (Virtual Private Cloud) Endpoint Service Private DNS Verification.
---
# Resource: aws_vpc_endpoint_service_private_dns_verification

Terraform resource for managing an AWS VPC (Virtual Private Cloud) Endpoint Service Private DNS Verification.
This resource begins the verification process by calling the [`StartVpcEndpointServicePrivateDnsVerification`](https://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_StartVpcEndpointServicePrivateDnsVerification.html) API.
The service provider should add a record to the DNS server _before_ creating this resource.

For additional details, refer to the AWS documentation on [managing VPC endpoint service DNS names](https://docs.aws.amazon.com/vpc/latest/privatelink/manage-dns-names.html).

~> Destruction of this resource will not stop the verification process, only remove the resource from state.

## Example Usage

### Basic Usage

```terraform
resource "aws_vpc_endpoint_service_private_dns_verification" "example" {
  service_id = aws_vpc_endpoint_service.example.id
}
```

## Argument Reference

The following arguments are required:

* `service_id` - (Required) ID of the endpoint service.

The following arguments are optional:

* `wait_for_verification` - (Optional) Whether to wait until the endpoint service returns a `Verified` status for the configured private DNS name.

## Attribute Reference

This resource exports no additional attributes.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `30m`)

## Import

You cannot import this resource.
