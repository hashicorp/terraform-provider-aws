---
layout: "aws"
page_title: "AWS: aws_worklink_fleet"
sidebar_current: "docs-aws-resource-worklink-fleet"
description: |-
  Provides a AWS WorkLink Fleet resource.
---

# aws_worklink_fleet

## Example Usage

Basic usage:

```hcl
resource "aws_worklink_fleet" "example" {
  name  = "terraform-example"
}
```

Network Configuration Usage:

```hcl
resource "aws_worklink_fleet" "example" {
  name = "terraform-example"

  network {
		vpc_id = "${aws_vpc.test.id}"
		subnet_ids = ["${aws_subnet.test.*.id}"]
		security_group_ids = ["${aws_security_group.test.id}"]
  }
}
```

Identity Provider Configuration Usage:

```hcl
resource "aws_worklink_fleet" "test" {
	name = "tf-worklink-fleet-%s"

	identity_provider {
		type = "SAML"
		saml_metadata = "${file("saml-metadata.xml")}"
	}
}
```


## Argument Reference

The following arguments are supported:

* `name` - (Required) A region-unique name for the AMI.
* `display_name` - (Optional) The name of the fleet.
* `optimize_for_end_user_location` - (Optional) The option to optimize for better performance by routing traffic through the closest AWS Region to users, which may be outside of your home Region. Defaults to `true`.
* `audit_stream_arn` - (Optional) The ARN of the Amazon Kinesis data stream that receives the audit events.
* `network` - (Optional) Provide this to allow manage the company network configuration for the fleet. Fields documented below.
* `device_ca_certificate` - (Optional) The certificate chain, including intermediate certificates and the root certificate authority certificate used to issue device certificates.
* `identity_provider` - (Optional) Provide this to allow manage the identity provider configuration for the fleet. Fields documented below.

**network** requires the following:

* `vpc_id` - (Required) The VPC ID with connectivity to associated websites.
* `subnet_ids` - (Required) A list of subnet IDs used for X-ENI connections from Amazon WorkLink rendering containers.
* `security_group_ids` - (Required) A list of security group IDs associated with access to the provided subnets.

**identity_provider** requires the following:

* `type` - (Required) The type of identity provider.
* `saml_metadata` - (Optional) The SAML metadata document provided by the customer’s identity provider. The existing IdentityProviderSamlMetadata is unset if null is passed.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The ARN of the created WorkLink Fleet.
* `arn` - The ARN of the created WorkLink Fleet.
* `company_code` - The identifier used by users to sign in to the Amazon WorkLink app.
* `created_time` - The time that the fleet was created.
* `last_updated_time` - The time that the fleet was last updated.

## Import

WorkLink can be imported using the ARN, e.g.

```
$ terraform import aws_worklink_fleet.test arn:aws:worklink::123456789012:fleet/example
```