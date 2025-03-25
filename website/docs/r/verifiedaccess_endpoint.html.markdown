---
subcategory: "Verified Access"
layout: "aws"
page_title: "AWS: aws_verifiedaccess_endpoint"
description: |-
  Terraform resource for managing a Verified Access Endpoint.
---

# Resource: aws_verifiedaccess_endpoint

Terraform resource for managing an AWS EC2 (Elastic Compute Cloud) Verified Access Endpoint.

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

* `attachment_type` - (Required) The type of attachment. Currently, only `vpc` is supported.
* `endpoint_domain_prefix` - (Required) - A custom identifier that is prepended to the DNS name that is generated for the endpoint.
* `endpoint_type` - (Required) - The type of Verified Access endpoint to create. Currently `load-balancer` or `network-interface` are supported.
* `verified_access_group_id` (Required) - The ID of the Verified Access group to associate the endpoint with.

The following arguments are optional:

* `application_domain` - (Optional) The DNS name for users to reach your application. This parameter is required if the endpoint type is `load-balancer` or `network-interface`.
* `description` - (Optional) A description for the Verified Access endpoint.
* `domain_certificate_arn` - (Optional) - The ARN of the public TLS/SSL certificate in AWS Certificate Manager to associate with the endpoint. The CN in the certificate must match the DNS name your end users will use to reach your application. This parameter is required if the endpoint type is `load-balancer` or `network-interface`.
* `sse_specification` - (Optional) The options in use for server side encryption.
* `load_balancer_options` - (Optional) The load balancer details. This parameter is required if the endpoint type is `load-balancer`.
* `network_interface_options` - (Optional) The network interface details. This parameter is required if the endpoint type is `network-interface`.
* `cidr_options` - (Optional) The CIDR block details. This parameter is required if the endpoint type is `cidr`.
* `policy_document` - (Optional) The policy document that is associated with this resource.
* `security_group_ids` - (Optional) List of the the security groups IDs to associate with the Verified Access endpoint.
* `tags` - (Optional) Key-value tags for the Verified Access Endpoint. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `device_validation_domain` - Returned if endpoint has a device trust provider attached.
* `endpoint_domain` - A DNS name that is generated for the endpoint.
* `id` - The ID of the AWS Verified Access endpoint.

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
