---
subcategory: "Lambda Web"
layout: "aws"
page_title: "AWS: aws_lambdaweb_function"
description: |-
  Provides details about an AWS Lambda Web function.
---

<!-- Copyright IBM Corp. 2014, 2026 -->
<!-- SPDX-License-Identifier: MPL-2.0 -->

# Data Source: aws_lambdaweb_function

Provides details about an AWS Lambda Web function.

## Example Usage

```terraform
data "aws_lambdaweb_function" "example" {
  function_name = "example"
}
```

## Argument Reference

The following arguments are required:

* `function_name` - (Required) Name of the function.

The following arguments are optional:

* `region` - (Optional) Region where this resource will be managed. Defaults to the Region set in the provider configuration.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `arn` - ARN of the function.
* `latest_revision_id` - ID of the newest revision of the function.
* `state` - Current state of the function.
* `state_reason` - Reason for the current state.
