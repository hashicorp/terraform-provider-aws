---
subcategory: "Directory Service"
layout: "aws"
page_title: "AWS: aws_directory_service_radius_settings"
description: |-
  Manages a directory's multi-factor authentication (MFA) using a Remote Authentication Dial In User Service (RADIUS) server.
---

# Resource: aws_directory_service_radius_settings

Manages a directory's multi-factor authentication (MFA) using a Remote Authentication Dial In User Service (RADIUS) server.

## Example Usage

```terraform
resource "aws_directory_service_radius_settings" "example" {
  directory_id = aws_directory_service_directory.example.id

  authentication_protocol = "PAP"
  display_label           = "example"
  radius_port             = 1812
  radius_retries          = 4
  radius_servers          = ["10.0.1.5"]
  radius_timeout          = 1
  shared_secret           = "12345678"
}
```

## Argument Reference

This resource supports the following arguments:

* `authentication_protocol` - (Optional) The protocol specified for your RADIUS endpoints. Valid values: `PAP`, `CHAP`, `MS-CHAPv1`, `MS-CHAPv2`.
* `directory_id` - (Required) The identifier of the directory for which you want to manager RADIUS settings.
* `display_label` - (Required) Display label.
* `radius_port` - (Required) The port that your RADIUS server is using for communications. Your self-managed network must allow inbound traffic over this port from the AWS Directory Service servers.
* `radius_retries` - (Required) The maximum number of times that communication with the RADIUS server is attempted. Minimum value of `0`. Maximum value of `10`.
* `radius_servers` - (Required) An array of strings that contains the fully qualified domain name (FQDN) or IP addresses of the RADIUS server endpoints, or the FQDN or IP addresses of your RADIUS server load balancer.
* `radius_timeout` - (Required) The amount of time, in seconds, to wait for the RADIUS server to respond. Minimum value of `1`. Maximum value of `50`.
* `shared_secret` - (Required) Required for enabling RADIUS on the directory.
* `use_same_username` - (Optional) Not currently used.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - The directory identifier.

## Timeouts

`aws_directory_service_radius_settings` provides the following [Timeouts](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts) configuration options:

- `create` - (Default `30 minutes`) Used for RADIUS settings creation
- `update` - (Default `30 minutes`) Used for RADIUS settings update

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import RADIUS settings using the directory ID. For example:

```terraform
import {
  to = aws_directory_service_radius_settings.example
  id = "d-926724cf57"
}
```

Using `terraform import`, import RADIUS settings using the directory ID. For example:

```console
% terraform import aws_directory_service_radius_settings.example d-926724cf57
```
