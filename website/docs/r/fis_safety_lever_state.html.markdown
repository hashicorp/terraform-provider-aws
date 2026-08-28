---
subcategory: "FIS (Fault Injection Simulator)"
layout: "aws"
page_title: "AWS: aws_fis_safety_lever_state"
description: |-
  Manages the state of an AWS FIS (Fault Injection Simulator) account safety lever.
---

# Resource: aws_fis_safety_lever_state

Manages the state of the AWS FIS (Fault Injection Simulator) safety lever for the account and Region. The safety lever is a single, account/Region-wide emergency stop: engaging it immediately stops all running experiments and blocks new ones from starting.

There is exactly one safety lever per account and Region, and it always exists — AWS does not provide APIs to create or delete it. Because of this, deleting this resource only removes it from Terraform state; it does not change the live value in AWS, so it never risks silently disengaging a safety control.

~> **Note:** AWS rejects a `reason` change unless `status` is also actually transitioning (for example, `engaged` to `disengaged`). This means:
>
> * Creating this resource fails with an error if the safety lever's live status already matches the configured `status`. Use [`terraform import`](#import) instead to bring an existing safety lever under management without changing its status.
> * Changing only `reason` while `status` stays the same fails at apply time. Pair a `reason` change with an actual `status` change.

## Example Usage

```terraform
resource "aws_fis_safety_lever_state" "example" {
  state {
    status = "disengaged"
    reason = "Managed by Terraform"
  }
}
```

## Argument Reference

The following arguments are required:

* `state` - (Required) State of the safety lever. [See below](#state).

The following arguments are optional:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

### `state`

* `reason` - (Required) Reason for the current status of the safety lever.
* `status` - (Required) Status of the safety lever. Valid values: `engaged`, `disengaged`. Engaging the lever immediately stops all running experiments in the account and Region, and prevents new ones from starting.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the safety lever.
* `id` - ID of the safety lever. Always `default`.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `5m`)
* `update` - (Default `5m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import the FIS safety lever state using `default`. For example:

```terraform
import {
  to = aws_fis_safety_lever_state.example
  id = "default"
}
```

Using `terraform import`, import the FIS safety lever state using `default`. For example:

```console
% terraform import aws_fis_safety_lever_state.example default
```
