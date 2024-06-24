---
subcategory: "SWF (Simple Workflow)"
layout: "aws"
page_title: "AWS: aws_swf_domain"
description: |-
  Provides an SWF Domain resource
---

# Resource: aws_swf_domain

Provides an SWF Domain resource.

## Example Usage

To register a basic SWF domain:

```terraform
resource "aws_swf_domain" "foo" {
  name                                        = "foo"
  description                                 = "Terraform SWF Domain"
  workflow_execution_retention_period_in_days = 30
}
```

## Argument Reference

This resource supports the following arguments:

* `name` - (Optional, Forces new resource) The name of the domain. If omitted, Terraform will assign a random, unique name.
* `name_prefix` - (Optional, Forces new resource) Creates a unique name beginning with the specified prefix. Conflicts with `name`.
* `description` - (Optional, Forces new resource) The domain description.
* `workflow_execution_retention_period_in_days` - (Required, Forces new resource) Length of time that SWF will continue to retain information about the workflow execution after the workflow execution is complete, must be between 0 and 90 days.
* `tags` - (Optional) Key-value map of resource tags. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - The name of the domain.
* `arn` - Amazon Resource Name (ARN)
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import SWF Domains using the `name`. For example:

```terraform
import {
  to = aws_swf_domain.foo
  id = "test-domain"
}
```

Using `terraform import`, import SWF Domains using the `name`. For example:

```console
% terraform import aws_swf_domain.foo test-domain
```
