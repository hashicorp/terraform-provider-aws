---
subcategory: "VPC (Virtual Private Cloud)"
layout: "aws"
page_title: "AWS: aws_vpc_endpoint_service"
description: |-
    Provides details about a specific service that can be specified when creating a VPC endpoint.
---

# Data Source: aws_vpc_endpoint_service

The VPC Endpoint Service data source details about a specific service that
can be specified when creating a VPC endpoint within the region configured in the provider.

## Example Usage

### AWS Service

```terraform
# Declare the data source
data "aws_vpc_endpoint_service" "s3" {
  service      = "s3"
  service_type = "Gateway"
}

# Create a VPC
resource "aws_vpc" "foo" {
  cidr_block = "10.0.0.0/16"
}

# Create a VPC endpoint
resource "aws_vpc_endpoint" "ep" {
  vpc_id       = aws_vpc.foo.id
  service_name = data.aws_vpc_endpoint_service.s3.service_name
}
```

### Non-AWS Service

```terraform
data "aws_vpc_endpoint_service" "custome" {
  service_name = "com.amazonaws.vpce.us-west-2.vpce-svc-0e87519c997c63cd8"
}
```

### Filter

```terraform
data "aws_vpc_endpoint_service" "test" {
  filter {
    name   = "service-name"
    values = ["some-service"]
  }
}
```

## Argument Reference

The arguments of this data source act as filters for querying the available VPC endpoint services.
The given filters must match exactly one VPC endpoint service whose data will be exported as attributes.

* `filter` - (Optional) Configuration block(s) for filtering. Detailed below.
* `service` - (Optional) Common name of an AWS service (e.g., `s3`).
* `service_name` - (Optional) Service name that is specified when creating a VPC endpoint. For AWS services the service name is usually in the form `com.amazonaws.<region>.<service>` (the SageMaker Notebook service is an exception to this rule, the service name is in the form `aws.sagemaker.<region>.notebook`).
* `service_type` - (Optional) Service type, `Gateway` or `Interface`.
* `tags` - (Optional) Map of tags, each pair of which must exactly match a pair on the desired VPC Endpoint Service.

~> **NOTE:** Specifying `service` will not work for non-AWS services or AWS services that don't follow the standard `service_name` pattern of `com.amazonaws.<region>.<service>`.

### filter Configuration Block

The following arguments are supported by the `filter` configuration block:

* `name` - (Required) Name of the filter field. Valid values can be found in the [EC2 DescribeVpcEndpointServices API Reference](https://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_DescribeVpcEndpointServices.html).
* `values` - (Required) Set of values that are accepted for the given filter field. Results will be selected if any given value matches.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `acceptance_required` - Whether or not VPC endpoint connection requests to the service must be accepted by the service owner - `true` or `false`.
* `arn` - ARN of the VPC endpoint service.
* `availability_zones` - Availability Zones in which the service is available.
* `base_endpoint_dns_names` - The DNS names for the service.
* `manages_vpc_endpoints` - Whether or not the service manages its VPC endpoints - `true` or `false`.
* `owner` - AWS account ID of the service owner or `amazon`.
* `private_dns_name` - Private DNS name for the service.
* `service_id` - ID of the endpoint service.
* `supported_ip_address_types` - The supported IP address types.
* `tags` - Map of tags assigned to the resource.
* `vpc_endpoint_policy_supported` - Whether or not the service supports endpoint policies - `true` or `false`.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

- `read` - (Default `20m`)
