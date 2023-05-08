---
subcategory: "Connect"
layout: "aws"
page_title: "AWS: aws_connect_phone_number"
description: |-
  Provides details about a specific Amazon Connect Phone Number.
---

# Resource: aws_connect_phone_number

Provides an Amazon Connect Phone Number resource. For more information see
[Amazon Connect: Getting Started](https://docs.aws.amazon.com/connect/latest/adminguide/amazon-connect-get-started.html)

## Example Usage

### Basic

```terraform
resource "aws_connect_phone_number" "example" {
  target_arn   = aws_connect_instance.example.arn
  country_code = "US"
  type         = "DID"

  tags = {
    "hello" = "world"
  }
}
```

### Description

```terraform
resource "aws_connect_phone_number" "example" {
  target_arn   = aws_connect_instance.example.arn
  country_code = "US"
  type         = "DID"
  description  = "example description"
}
```

### Prefix to filter phone numbers

```terraform
resource "aws_connect_phone_number" "example" {
  target_arn   = aws_connect_instance.example.arn
  country_code = "US"
  type         = "DID"
  prefix       = "+18005"
}
```

## Argument Reference

The following arguments are supported:

* `country_code` - (Required, Forces new resource) The ISO country code. For a list of Valid values, refer to [PhoneNumberCountryCode](https://docs.aws.amazon.com/connect/latest/APIReference/API_SearchAvailablePhoneNumbers.html#connect-SearchAvailablePhoneNumbers-request-PhoneNumberCountryCode).
* `description` - (Optional, Forces new resource) The description of the phone number.
* `prefix` - (Optional, Forces new resource) The prefix of the phone number that is used to filter available phone numbers. If provided, it must contain `+` as part of the country code. Do not specify this argument when importing the resource.
* `tags` - (Optional) Tags to apply to the Phone Number. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.
* `target_arn` - (Required) The Amazon Resource Name (ARN) for Amazon Connect instances that phone numbers are claimed to.
* `type` - (Required, Forces new resource) The type of phone number. Valid Values: `TOLL_FREE` | `DID`.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - The ARN of the phone number.
* `phone_number` - The phone number. Phone numbers are formatted `[+] [country code] [subscriber number including area code]`.
* `id` - The identifier of the phone number.
* `status` - A block that specifies status of the phone number. [Documented below](#status).
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

### `status`

The `status` configuration block supports the following attributes:

* `message` - The status message.
* `status` - The status of the phone number. Valid Values: `CLAIMED` | `IN_PROGRESS` | `FAILED`.

## Timeouts

[Configuration options](https://www.terraform.io/docs/configuration/blocks/resources/syntax.html#operation-timeouts):

* `create` - (Default `2m`)
* `update` - (Default `2m`)
* `delete` - (Default `2m`)

## Import

Amazon Connect Phone Numbers can be imported using its `id` e.g.,

```
$ terraform import aws_connect_phone_number.example 12345678-abcd-1234-efgh-9876543210ab
```
