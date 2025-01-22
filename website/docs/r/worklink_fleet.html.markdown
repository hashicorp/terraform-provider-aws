---
subcategory: "WorkLink"
layout: "aws"
page_title: "AWS: aws_worklink_fleet"
description: |-
  Provides a AWS WorkLink Fleet resource.
---

# Resource: aws_worklink_fleet

Provides a AWS WorkLink Fleet resource.

!> **WARNING:** The `aws_worklink_fleet` resource has been deprecated and will be removed in a future version. Use Amazon WorkSpaces Secure Browser instead.

## Example Usage

Basic usage:

```terraform
resource "aws_worklink_fleet" "example" {
  name = "terraform-example"
}
```

Network Configuration Usage:

```terraform
resource "aws_worklink_fleet" "example" {
  name = "terraform-example"

  network {
    vpc_id             = aws_vpc.test.id
    subnet_ids         = [aws_subnet.test[*].id]
    security_group_ids = [aws_security_group.test.id]
  }
}
```

Identity Provider Configuration Usage:

```terraform
resource "aws_worklink_fleet" "test" {
  name = "tf-worklink-fleet"

  identity_provider {
    type          = "SAML"
    saml_metadata = file("saml-metadata.xml")
  }
}
```

## Argument Reference

This resource supports the following arguments:

* `name` - (Required) A region-unique name for the AMI.
* `audit_stream_arn` - (Optional) The ARN of the Amazon Kinesis data stream that receives the audit events. Kinesis data stream name must begin with `"AmazonWorkLink-"`.
* `device_ca_certificate` - (Optional) The certificate chain, including intermediate certificates and the root certificate authority certificate used to issue device certificates.
* `identity_provider` - (Optional) Provide this to allow manage the identity provider configuration for the fleet. Fields documented below.
* `display_name` - (Optional) The name of the fleet.
* `network` - (Optional) Provide this to allow manage the company network configuration for the fleet. Fields documented below.
* `optimize_for_end_user_location` - (Optional) The option to optimize for better performance by routing traffic through the closest AWS Region to users, which may be outside of your home Region. Defaults to `true`.

**network** requires the following:

~> **NOTE:** `network` cannot be removed without force recreating by `terraform taint`.

* `vpc_id` - (Required) The VPC ID with connectivity to associated websites.
* `subnet_ids` - (Required) A list of subnet IDs used for X-ENI connections from Amazon WorkLink rendering containers.
* `security_group_ids` - (Required) A list of security group IDs associated with access to the provided subnets.

**identity_provider** requires the following:

~> **NOTE:** `identity_provider` cannot be removed without force recreating by `terraform taint`.

* `type` - (Required) The type of identity provider.
* `saml_metadata` - (Required) The SAML metadata document provided by the customer’s identity provider.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - The ARN of the created WorkLink Fleet.
* `arn` - The ARN of the created WorkLink Fleet.
* `company_code` - The identifier used by users to sign in to the Amazon WorkLink app.
* `created_time` - The time that the fleet was created.
* `last_updated_time` - The time that the fleet was last updated.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import WorkLink using the ARN. For example:

```terraform
import {
  to = aws_worklink_fleet.test
  id = "arn:aws:worklink::123456789012:fleet/example"
}
```

Using `terraform import`, import WorkLink using the ARN. For example:

```console
% terraform import aws_worklink_fleet.test arn:aws:worklink::123456789012:fleet/example
```
