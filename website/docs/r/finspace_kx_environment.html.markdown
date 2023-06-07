---
subcategory: "FinSpace"
layout: "aws"
page_title: "AWS: aws_finspace_kx_environment"
description: |-
  Terraform resource for managing an AWS FinSpace Kx Environment.
---

# Resource: aws_finspace_kx_environment

Terraform resource for managing an AWS FinSpace Kx Environment.

## Example Usage

### Basic Usage

```terraform
resource "aws_kms_key" "example" {
  description             = "Sample KMS Key"
  deletion_window_in_days = 7
}

resource "aws_finspace_kx_environment" "example" {
  name       = "my-tf-kx-environment"
  kms_key_id = aws_kms_key.example.arn
}
```

### With Network Setup

```terraform
resource "aws_kms_key" "example" {
  description             = "Sample KMS Key"
  deletion_window_in_days = 7
}

resource "aws_ec2_transit_gateway" "example" {
  description = "example"
}

resource "aws_finspace_kx_environment" "example_env" {
  name        = "my-tf-kx-environment"
  description = "Environment description"
  kms_key_id  = aws_kms_key.example.arn

  transit_gateway_configuration {
    transit_gateway_id  = aws_ec2_transit_gateway.example.id
    routable_cidr_space = "100.64.0.0/26"
  }

  custom_dns_configuration {
    custom_dns_server_name = "example.finspace.amazonaws.com"
    custom_dns_server_ip   = "10.0.0.76"
  }
}
```

## Argument Reference

The following arguments are required:

* `name` - (Required) Name of the KX environment that you want to create.
* `kms_key_id` - (Required) KMS key ID to encrypt your data in the FinSpace environment.

The following arguments are optional:

* `description` - (Optional) Description for the KX environment.
* `transit_gateway_configuration` - (Optional) Transit gateway and network configuration that is used to connect the KX environment to an internal network. Defined below.
* `custom_dns_configuration` - (Optional) List of DNS server name and server IP. This is used to set up Route-53 outbound resolvers. Defined below.
* `tags` - (Optional) List of key-value pairs to label the KX environment.

### transit_gateway_configuration

The transit_gateway_configuration block supports the following arguments:

* `transit_gateway_id` - (Required) Identifier of the transit gateway created by the customer to connect outbound traffics from KX network to your internal network.
* `routable_cidr_space` - (Required) Routing CIDR on behalf of KX environment. It could be any “/26 range in the 100.64.0.0 CIDR space. After providing, it will be added to the customer’s transit gateway routing table so that the traffics could be routed to KX network.

### custom_dns_configuration

The custom_dns_configuration block supports the following arguments:

* `custom_dns_server_name` - (Required) Name of the DNS server.
* `custom_dns_server_ip` - (Required) IP address of the DNS server.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - Amazon Resource Name (ARN) identifier of the KX environment.
* `id` - Unique identifier for the KX environment.
* `availability_zones` - AWS Availability Zone IDs that this environment is available in. Important when selecting VPC subnets to use in cluster creation.
* `infrastructure_account_id` - Unique identifier for the AWS environment infrastructure account.
* `created_timestamp` - Timestamp at which the environment is created in FinSpace. Value determined as epoch time in seconds. For example, the value for Monday, November 1, 2021 12:00:00 PM UTC is specified as 1635768000.
* `last_modified_timestamp` - Last timestamp at which the environment was updated in FinSpace. Value determined as epoch time in seconds. For example, the value for Monday, November 1, 2021 12:00:00 PM UTC is specified as 1635768000.
* `status` - Status of environment creation
* `tags_all` - Map of tags assigned to the resource.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `30m`)
* `update` - (Default `30m`)
* `delete` - (Default `30m`)

## Import

An AWS FinSpace Kx Environment can be imported using the `id`, e.g.,

```
$ terraform import aws_finspace_kx_environment.example n3ceo7wqxoxcti5tujqwzs
```
