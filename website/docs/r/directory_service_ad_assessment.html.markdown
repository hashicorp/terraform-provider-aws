---
subcategory: "Directory Service"
layout: "aws"
page_title: "AWS: aws_directory_service_directory"
description: |-
  Manages an AWS Directory Service AD Assessment.
---

# Resource: aws_directory_service_directory

Manages an AWS Directory Service AD Assessment, required for creation of AWS Managed AD (Hybrid Edition)

~> **Note:** The `self_managed_instance_ids` must have a one-to-one correspondence with `customer_dns_ips`, meaning that if the IP address for instance `i-10243410` is `10.24.34.100` and the IP address for instance `i-10243420` is `10.24.34.200`, then the inputs arrays must maintain the same order relationship, either `["10.24.34.100", "10.24.34.200"]` paired with `["i-10243410", "i-10243420"]` or `["10.24.34.200", "10.24.34.100"]` paired with `["i-10243420", "i-10243410"]`.

## Example Usage

### Basic Usage

```terraform
resource "aws_directory_service_ad_assessment" "test" {
  customer_dns_ips = ["10.24.34.100", "10.24.34.200"]
  dns_name         = "corp.notexample.com"
  self_managed_instance_ids = [
    "i-10243410",
    "i-10243420",
  ]
  security_group_ids = [aws_security_group.test.id]
  subnet_ids         = [aws_subnet.foo.id, aws_subnet.bar.id]
  vpc_id             = aws_vpc.main.id
}
```

## Argument Reference

The following arguments are required:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `customer_dns_ips` - (Required) List of IP addresses for the DNS servers or domain controllers in your self-managed AD that are tested during the assessment.
* `dns_name` - (Required) [FQDN](https://en.wikipedia.org/wiki/Fully_qualified_domain_name) of the self-managed AD domain to assess.
* `self_managed_instance_ids` - (Required) IDs of the self-managed instances with SSM that are used to perform connectivity and validation tests.
* `subnet_ids` - (Required) IDs of the subnets in which you intend to deploy directory servers (2 subnets in 2 different AZs).
* `vpc_id` - (Required) Identifier of the VPC in which  you intend to deploy directory servers

The following arguments are optional:

* `security_group_ids` - (Optional) List with exactly one security group id that allows network traffic to and from your self-managed domain controllers outside of your Amazon VPC. By default, the service attaches a security group to allow network access to the self-managed nodes in your Amazon VPC.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `assessment_id` - ID of the AD Assessment.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `45m`)
* `update` - (Default `30m`)
* `delete` - (Default `30m`)

## Import

In Terraform v1.12.0 and later, the [`import` block](https://developer.hashicorp.com/terraform/language/import) can be used with the `identity` attribute. For example:

```terraform
import {
  to = aws_directory_service_ad_assessment.example
  identity = {
    assessment_id = "da-12345678"
  }
}

resource "aws_directory_service_ad_assessment" "example" {
  ### Configuration omitted for brevity ###
}
```

### Identity Schema

#### Required

* `assessment_id` - ID argument of the AD Assessment.

#### Optional

* `account_id` (String) AWS Account where this resource is managed.
* `region` (String) Region where this resource is managed.

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Directory Service AD Assessment using the `assessment_id`. For example:

```terraform
import {
  to = aws_directory_service_ad_assessment.example
  id = "da-12345678"
}
```

Using `terraform import`, import Directory Service AD Assessment using the `assessment_id`. For example:

```console
% terraform import aws_directory_service_ad_assessment.example da-12345678
```
