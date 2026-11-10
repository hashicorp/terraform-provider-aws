---
subcategory: "Directory Service"
layout: "aws"
page_title: "AWS: aws_directory_service_directory_setting"
description: |-
  Manages a single setting for an AWS Directory Service directory.
---

# Resource: aws_directory_service_directory_setting

Manages a single configurable setting for an AWS Directory Service directory, such as enabling or disabling a TLS protocol version.

## Example Usage

### Basic Usage

```terraform
resource "aws_directory_service_directory_setting" "example" {
  directory_id = aws_directory_service_directory.example.id
  name         = "TLS_1_0"
  value        = "Disable"
}

resource "aws_directory_service_directory" "example" {
  name     = "corp.example.com"
  password = "SuperSecretPassw0rd"
  type     = "MicrosoftAD"

  vpc_settings {
    vpc_id     = aws_vpc.example.id
    subnet_ids = aws_subnet.example[*].id
  }
}
```

## Argument Reference

The following arguments are required:

* `directory_id` - (Required, Forces new resource) ID of the directory.
* `name` - (Required, Forces new resource) Name of the directory setting. For example, `TLS_1_0`.
* `value` - (Required) Value of the directory setting. For example, `Disable` or `Enable`.

The following arguments are optional:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `type` - Type of the directory setting.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `30m`)
* `update` - (Default `30m`)

## Import

In Terraform v1.12.0 and later, the [`import` block](https://developer.hashicorp.com/terraform/language/import) can be used with the `identity` attribute. For example:

```terraform
import {
  to = aws_directory_service_directory_setting.example
  identity = {
    directory_id = "d-1234567890"
    name         = "TLS_1_0"
  }
}

resource "aws_directory_service_directory_setting" "example" {
  ### Configuration omitted for brevity ###
}
```

### Identity Schema

#### Required

* `directory_id` (String) ID of the directory.
* `name` (String) Name of the directory setting.

#### Optional

* `account_id` (String) AWS Account where this resource is managed.
* `region` (String) Region where this resource is managed.

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import a directory setting using `directory_id,name`. For example:

```terraform
import {
  to = aws_directory_service_directory_setting.example
  id = "d-1234567890,TLS_1_0"
}
```

Using `terraform import`, import a directory setting using `directory_id,name`. For example:

```console
% terraform import aws_directory_service_directory_setting.example d-1234567890,TLS_1_0
```
