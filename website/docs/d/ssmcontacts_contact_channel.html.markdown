---
subcategory: "SSM Contacts"
layout: "aws"
page_title: "AWS: aws_ssmcontacts_contact_channel"
description: |-
  Terraform data source for managing an AWS SSM Contacts Contact Channel.
---

# Data Source: aws_ssmcontacts_contact_channel

Terraform data source for managing an AWS SSM Contacts Contact Channel.

## Example Usage

### Basic Usage

```terraform
data "aws_ssmcontacts_contact_channel" "example" {
  arn = "arn:aws:ssm-contacts:us-west-2:123456789012:contact-channel/example"
}
```

## Argument Reference

This data source supports the following arguments:

- `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
- `arn` - Amazon Resource Name (ARN) of the contact channel.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

- `activation_status` - Whether the contact channel is activated.
- `contact_id` - Amazon Resource Name (ARN) of the AWS SSM Contact that the contact channel belongs to.
- `delivery_address` - Details used to engage the contact channel.
- `name` - Name of the contact channel.
- `type` - Type of the contact channel.
