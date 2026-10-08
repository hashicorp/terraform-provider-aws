---
subcategory: "Verified Access"
layout: "aws"
page_title: "AWS: aws_verifiedaccess_endpoint"
description: |-
  Terraform resource for managing a Verified Access Endpoint.
---

# Resource: aws_verifiedaccess_endpoint

Terraform resource for managing an AWS EC2 Verified Access Endpoint.

## Example Usage

### ALB Example

```terraform
resource "aws_verifiedaccess_endpoint" "example" {
  application_domain     = "example.com"
  attachment_type        = "vpc"
  description            = "example"
  domain_certificate_arn = aws_acm_certificate.example.arn
  endpoint_domain_prefix = "example"
  endpoint_type          = "load-balancer"
  load_balancer_options {
    load_balancer_arn = aws_lb.example.arn
    port              = 443
    protocol          = "https"
    subnet_ids        = [for subnet in aws_subnet.public : subnet.id]
  }
  security_group_ids       = [aws_security_group.example.id]
  verified_access_group_id = aws_verifiedaccess_group.example.id
}
```

### Network Interface Example

```terraform
resource "aws_verifiedaccess_endpoint" "example" {
  application_domain     = "example.com"
  attachment_type        = "vpc"
  description            = "example"
  domain_certificate_arn = aws_acm_certificate.example.arn
  endpoint_domain_prefix = "example"
  endpoint_type          = "network-interface"
  network_interface_options {
    network_interface_id = aws_network_interface.example.id
    port                 = 443
    protocol             = "https"
  }
  security_group_ids       = [aws_security_group.example.id]
  verified_access_group_id = aws_verifiedaccess_group.example.id
}
```

### Cidr Example

```terraform
resource "aws_verifiedaccess_endpoint" "example" {
  attachment_type = "vpc"
  description     = "example"
  endpoint_type   = "cidr"
  cidr_options {
    cidr = aws_subnet.test[0].cidr_block
    port_range {
      from_port = 443
      to_port   = 443
    }
    protocol   = "tcp"
    subnet_ids = [for subnet in aws_subnet.test : subnet.id]
  }

  security_group_ids       = [aws_security_group.test.id]
  verified_access_group_id = aws_verifiedaccess_group.test.id
}
```

## Argument Reference

The following arguments are required:

* `attachment_type` - (Required) Type of attachment. Currently, only `vpc` is supported.
* `endpoint_type` - (Required) Type of Verified Access endpoint to create. Valid values are `load-balancer`, `network-interface`, `cidr`, and `rds`.
* `verified_access_group_id` - (Required) ID of the Verified Access group to associate the endpoint with.

The following arguments are optional:

* `application_domain` - (Optional) DNS name for users to reach your application. This parameter is required if the endpoint type is `load-balancer` or `network-interface`.
* `cidr_options` - (Optional) CIDR block details. This parameter is required if the endpoint type is `cidr`. [See below](#cidr_options-block).
* `description` - (Optional) Description for the Verified Access endpoint.
* `domain_certificate_arn` - (Optional) ARN of the public TLS/SSL certificate in AWS Certificate Manager to associate with the endpoint. The CN in the certificate must match the DNS name your end users will use to reach your application. This parameter is required if the endpoint type is `load-balancer` or `network-interface`.
* `endpoint_domain_prefix` - (Optional) Custom identifier that is prepended to the DNS name that is generated for the endpoint.
* `load_balancer_options` - (Optional) Load balancer details. This parameter is required if the endpoint type is `load-balancer`. [See below](#load_balancer_options-block).
* `network_interface_options` - (Optional) Network interface details. This parameter is required if the endpoint type is `network-interface`. [See below](#network_interface_options-block).
* `policy_document` - (Optional) Policy document that is associated with this resource.
* `rds_options` - (Optional) RDS details. This parameter is required if the endpoint type is `rds`. [See below](#rds_options-block).
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `security_group_ids` - (Optional) List of the security groups IDs to associate with the Verified Access endpoint.
* `sse_specification` - (Optional) Options in use for server side encryption. [See below](#sse_specification-block).
* `tags` - (Optional) Key-value tags for the Verified Access Endpoint. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

### `cidr_options` Block

* `cidr` - (Required) CIDR block to send traffic to.
* `port_range` - (Required) Port ranges. [See below](#cidr_optionsport_range-block).
* `protocol` - (Optional) Protocol. Currently `tcp` is supported.
* `subnet_ids` - (Optional) IDs of the subnets.

### `cidr_options.port_range` Block

* `from_port` - (Required) Start of the port range.
* `to_port` - (Required) End of the port range.

### `load_balancer_options` Block

* `load_balancer_arn` - (Optional) ARN of the load balancer.
* `port` - (Optional) IP port number.
* `port_range` - (Optional) Port ranges. [See below](#load_balancer_optionsport_range-block).
* `protocol` - (Optional) IP protocol.
* `subnet_ids` - (Optional) IDs of the subnets.

### `load_balancer_options.port_range` Block

* `from_port` - (Required) Start of the port range.
* `to_port` - (Required) End of the port range.

### `network_interface_options` Block

* `network_interface_id` - (Optional) ID of the network interface.
* `port` - (Optional) IP port number.
* `port_range` - (Optional) Port ranges. [See below](#network_interface_optionsport_range-block).
* `protocol` - (Optional) IP protocol.

### `network_interface_options.port_range` Block

* `from_port` - (Required) Start of the port range.
* `to_port` - (Required) End of the port range.

### `rds_options` Block

* `port` - (Optional) IP port number.
* `protocol` - (Optional) Protocol. Currently `tcp` is supported.
* `rds_db_cluster_arn` - (Optional) ARN of the RDS cluster.
* `rds_db_instance_arn` - (Optional) ARN of the RDS instance.
* `rds_db_proxy_arn` - (Optional) ARN of the RDS proxy.
* `rds_endpoint` - (Optional) RDS endpoint.
* `subnet_ids` - (Optional) IDs of the subnets.

### `sse_specification` Block

* `customer_managed_key_enabled` - (Optional) Whether to encrypt the policy using a customer managed key.
* `kms_key_arn` - (Optional) ARN of the KMS key.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `device_validation_domain` - Returned if endpoint has a device trust provider attached.
* `endpoint_domain` - DNS name that is generated for the endpoint.
* `id` - ID of the AWS Verified Access endpoint.
* `verified_access_instance_id` - ID of the Verified Access instance.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `60m`)
* `update` - (Default `180m`)
* `delete` - (Default `90m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Verified Access Instances using the `id`. For example:

```terraform
import {
  to = aws_verifiedaccess_endpoint.example
  id = "vae-8012925589"
}
```

Using `terraform import`, import Verified Access Instances using the  `id`. For example:

```console
% terraform import aws_verifiedaccess_endpoint.example vae-8012925589
```
