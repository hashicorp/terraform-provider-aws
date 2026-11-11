---
subcategory: "CloudWatch Omni"
layout: "aws"
page_title: "AWS: aws_cloudwatchomni_dashboard"
description: |-
  Manages an Amazon CloudWatch Omni dashboard.
---

# Resource: aws_cloudwatchomni_dashboard

Manages an Amazon CloudWatch Omni dashboard.

## Example Usage

```terraform
resource "aws_cloudwatchomni_dashboard" "example" {
  space_id    = aws_cloudwatchomni_space.example.space_id
  name        = "example"
  description = "Example dashboard"

  tags = {
    Environment = "production"
  }

  body = jsonencode({
    panels = [{
      type   = "markdown"
      layout = { x = 0, y = 0, w = 60, h = "auto" }
      config = { content = "## On call\n\nCheck the error rate first." }
    }]
  })
}
```

## Argument Reference

The following arguments are required:

* `body` - (Required) Dashboard definition, as a JSON document with a `panels` array. Maximum size of 1 MiB.
* `name` - (Required) Name of the dashboard. Must be 1-256 characters and can contain letters, numbers, and the characters `_`, `.`, `@`, `~`, `(`, `)`, and `-`.
* `space_id` - (Required, Forces new resource) ID of the space that contains the dashboard.

The following arguments are optional:

* `description` - (Optional) Description of the dashboard. Must be 1-1024 characters. Removing the description replaces the dashboard, because the service does not support clearing it.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `tags` - (Optional) Map of tags assigned to the resource. If configured with a provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the dashboard.
* `dashboard_id` - Unique identifier of the dashboard.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Import

In Terraform v1.12.0 and later, you can use an [`import` block](https://developer.hashicorp.com/terraform/language/import) with the `identity` attribute. For example:

```terraform
import {
  to = aws_cloudwatchomni_dashboard.example
  identity = {
    space_id     = "12345678-1234-1234-1234-123456789012"
    dashboard_id = "87654321-4321-4321-4321-210987654321"
  }
}

resource "aws_cloudwatchomni_dashboard" "example" {
  ### Configuration omitted for brevity ###
}
```

### Identity Schema

#### Required

* `dashboard_id` (String) Dashboard ID.
* `space_id` (String) Space ID.

#### Optional

* `account_id` (String) AWS Account where this resource is managed.
* `region` (String) Region where this resource is managed.

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import a CloudWatch Omni Dashboard using the `space_id` and `dashboard_id` separated by a comma (`,`). For example:

```terraform
import {
  to = aws_cloudwatchomni_dashboard.example
  id = "12345678-1234-1234-1234-123456789012,87654321-4321-4321-4321-210987654321"
}
```

Using `terraform import`, import a CloudWatch Omni Dashboard using the `space_id` and `dashboard_id` separated by a comma (`,`). For example:

```console
% terraform import aws_cloudwatchomni_dashboard.example 12345678-1234-1234-1234-123456789012,87654321-4321-4321-4321-210987654321
```
