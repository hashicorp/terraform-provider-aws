---
subcategory: "EC2 (Elastic Compute Cloud)"
layout: "aws"
page_title: "AWS: aws_ec2_application_status_check"
description: |-
  Manages an EC2 Application Status Check.
---

# Resource: aws_ec2_application_status_check

Manages an EC2 Application Status Check. Application status checks monitor the reachability and response of an application endpoint.

## Example Usage

### Basic Usage

```hcl
resource "aws_ec2_application_status_check" "example" {
  protocol = "http"
  port     = 80
}
```

### Customer-Managed Health Check Path

```hcl
resource "aws_ec2_application_status_check" "example" {
  protocol = "https"
  port     = 443

  health_check_path {
    source {
      subnet_id         = "subnet-0123456789abcdef0"
      security_group_id = "sg-0123456789abcdef0"
    }

    destination {
      subnet_id         = "subnet-0123456789abcdef1"
      security_group_id = "sg-0123456789abcdef1"
    }
  }
}
```

## Argument Reference

The following arguments are required:

* `port` - (Required) Port on which the application accepts requests. Valid values are between `1` and `65535`.
* `protocol` - (Required) Protocol used to reach the application. Valid values are `http` and `https`.

The following arguments are optional:

* `aggregation` - (Optional) Whether this check is included in the aggregated instance status. Valid values are `included` and `excluded`. Defaults to `included`.
* `device_index` - (Optional) Network device index used for the health check. Must be greater than or equal to `0`. Defaults to `0`.
* `failure_threshold` - (Optional) Number of consecutive failed attempts before the application is considered unhealthy. Must be greater than `0`. Defaults to `2`.
* `health_check_path` - (Optional) Customer-managed network path used for the health check. See [`health_check_path` Block](#health_check_path-block) below.
* `initialization_grace_period_seconds` - (Optional) Number of seconds allowed for application initialization before health checks affect status. Valid values are between `1` and `600`. Defaults to `300`.
* `interval` - (Optional) Number of seconds between health checks. Valid value is `60`. Defaults to `60`.
* `ip_scope` - (Optional) IP scope used for the health check. Valid value is `private`. Defaults to `private`.
* `ip_version` - (Optional) IP version used for the health check. Valid values are `ipv4` and `ipv6`. Defaults to `ipv4`.
* `path` - (Optional) Request path used for HTTP and HTTPS health checks. Must start with `/`. Defaults to `/`.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `status_code_matcher` - (Optional) HTTP status codes that indicate a successful response. Length must be between `1` and `64` characters. Defaults to `200`.
* `success_threshold` - (Optional) Number of consecutive successful attempts before the application is considered healthy. Must be greater than `0`. Defaults to `2`.
* `tags` - (Optional) Map of tags assigned to the resource. If configured with a provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.
* `timeout` - (Optional) Number of seconds to wait for a health check response. Valid values are between `1` and `30`. Defaults to `6`.

### `health_check_path` Block

The `health_check_path` block supports the following arguments:

* `destination` - (Required) One or more destination network locations. See [`destination` Block](#destination-block) below.
* `source` - (Required) Source network location. Exactly one source block must be configured. See [`source` Block](#source-block) below.

#### `destination` Block

The `destination` block supports the following arguments:

* `security_group_id` - (Required) ID of the security group for the destination network interface.
* `subnet_id` - (Required) ID of the subnet for the destination network interface.

#### `source` Block

The `source` block supports the following arguments:

* `security_group_id` - (Required) ID of the security group for the source network interface.
* `subnet_id` - (Required) ID of the subnet for the source network interface.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - ID of the application status check.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Import

In Terraform v1.12.0 and later, the [`import` block](https://developer.hashicorp.com/terraform/language/import) can be used with the `identity` attribute. For example:

```terraform
import {
  to = aws_ec2_application_status_check.example
  identity = {
    id = "asc-0123456789abcdef0"
  }
}
```

### Identity Schema

#### Required

* `id` (String) ID of the application status check.

#### Optional

* `account_id` (String) AWS Account where this resource is managed.
* `region` (String) Region where this resource is managed.

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import EC2 Application Status Checks using the application status check ID. For example:

```terraform
import {
  to = aws_ec2_application_status_check.example
  id = "asc-0123456789abcdef0"
}
```

Using `terraform import`, import EC2 Application Status Checks using the application status check ID. For example:

```console
% terraform import aws_ec2_application_status_check.example asc-0123456789abcdef0
```
