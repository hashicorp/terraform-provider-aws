---
subcategory: "Directory Service"
layout: "aws"
page_title: "AWS: aws_directory_service_trust"
description: |-
  Manages a trust relationship between two Active Directory Directories.
---

# Resource: aws_directory_service_trust

Manages a trust relationship between two Active Directory Directories.

The directories may either be both AWS Managed Microsoft AD domains or an AWS Managed Microsoft AD domain and a self-managed Active Directory Domain.

The Trust relationship must be configured on both sides of the relationship.
If a Trust has only been created on one side, it will be in the state `VerifyFailed`.
Once the second Trust is created, the first will update to the correct state.

## Example Usage

### Two-Way Trust

```terraform
resource "aws_directory_service_trust" "one" {
  directory_id = aws_directory_service_directory.one.id

  remote_domain_name = aws_directory_service_directory.two.name
  trust_direction    = "Two-Way"
  trust_password     = "Some0therPassword"

  conditional_forwarder_ip_addrs = aws_directory_service_directory.two.dns_ip_addresses
}

resource "aws_directory_service_trust" "two" {
  directory_id = aws_directory_service_directory.two.id

  remote_domain_name = aws_directory_service_directory.one.name
  trust_direction    = "Two-Way"
  trust_password     = "Some0therPassword"

  conditional_forwarder_ip_addrs = aws_directory_service_directory.one.dns_ip_addresses
}

resource "aws_directory_service_directory" "one" {
  name = "one.example.com"
  type = "MicrosoftAD"
  # ...
}

resource "aws_directory_service_directory" "two" {
  name = "two.example.com"
  type = "MicrosoftAD"
  # ...
}
```

### One-Way Trust

```terraform
resource "aws_directory_service_trust" "one" {
  directory_id = aws_directory_service_directory.one.id

  remote_domain_name = aws_directory_service_directory.two.name
  trust_direction    = "One-Way: Incoming"
  trust_password     = "Some0therPassword"

  conditional_forwarder_ip_addrs = aws_directory_service_directory.two.dns_ip_addresses
}

resource "aws_directory_service_trust" "two" {
  directory_id = aws_directory_service_directory.two.id

  remote_domain_name = aws_directory_service_directory.one.name
  trust_direction    = "One-Way: Outgoing"
  trust_password     = "Some0therPassword"

  conditional_forwarder_ip_addrs = aws_directory_service_directory.one.dns_ip_addresses
}

resource "aws_directory_service_directory" "one" {
  name = "one.example.com"
  type = "MicrosoftAD"
  # ...
}

resource "aws_directory_service_directory" "two" {
  name = "two.example.com"
  type = "MicrosoftAD"
  # ...
}
```

## Argument Reference

This resource supports the following arguments:

* `conditional_forwarder_ip_addrs` - (Optional) Set of IPv4 addresses for the DNS server associated with the remote Directory.
  Can contain between 1 and 4 values.
* `delete_associated_conditional_forwarder` - (Optional) Whether to delete the conditional forwarder when deleting the Trust relationship.
* `directory_id` - (Required) ID of the Directory.
* `remote_domain_name` - (Required) Fully qualified domain name of the remote Directory.
* `selective_auth` - (Optional) Whether to enable selective authentication.
  Valid values are `Enabled` and `Disabled`.
  Default value is `Disabled`.
* `trust_direction` - (Required) The direction of the Trust relationship.
  Valid values are `One-Way: Outgoing`, `One-Way: Incoming`, and `Two-Way`.
* `trust_password` - (Required) Password for the Trust.
  Does not need to match the passwords for either Directory.
  Can contain upper- and lower-case letters, numbers, and punctuation characters.
  May be up to 128 characters long.
* `trust_type` - (Optional) Type of the Trust relationship.
  Valid values are `Forest` and `External`.
  Default value is `Forest`.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `created_date_time` - Date and time when the Trust was created.
* `id` - The Trust identifier.
* `last_updated_date_time` - Date and time when the Trust was last updated.
* `state_last_updated_date_time` - Date and time when the Trust state in `trust_state` was last updated.
* `trust_state` - State of the Trust relationship.
  One of `Created`, `VerifyFailed`,`Verified`, `UpdateFailed`,`Updated`,`Deleted`, or `Failed`.
* `trust_state_reason` - Reason for the Trust state set in `trust_state`.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import the Trust relationship using the directory ID and remote domain name, separated by a `/`. For example:

```terraform
import {
  to = aws_directory_service_trust.example
  id = "d-926724cf57/directory.example.com"
}
```

Using `terraform import`, import the Trust relationship using the directory ID and remote domain name, separated by a `/`. For example:

```console
% terraform import aws_directory_service_trust.example d-926724cf57/directory.example.com
```
