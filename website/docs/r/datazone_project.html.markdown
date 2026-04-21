---
subcategory: "DataZone"
layout: "aws"
page_title: "AWS: aws_datazone_project"
description: |-
  Terraform resource for managing an Amazon DataZone Project.
---
# Resource: aws_datazone_project

Terraform resource for managing an AWS DataZone Project.

## Example Usage

```terraform
resource "aws_datazone_project" "test" {
  domain_id           = aws_datazone_domain.test.id
  glossary_terms      = ["2N8w6XJCwZf"]
  name                = "name"
  description         = "desc"
  skip_deletion_check = true
}
```

### Basic Usage

```terraform
resource "aws_datazone_project" "test" {
  domain_identifier = aws_datazone_domain.test.id
  name              = "name"
}
```

## Argument Reference

The following arguments are required:

* `domain_identifier` - (Required) Identifier of domain which the project is part of. Must follow the regex of `^dzd[-_][a-zA-Z0-9_-]{1,36}$`.
* `name` - (Required) Name of the project. Must follow the regex of `^[\w -]+$`. and have a length of at most 64.

The following arguments are optional:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `skip_deletion_check` - (Optional) Optional flag to delete all child entities within the project.
* `description` - (Optional) Description of project.
* `glossary_terms` - (Optional) List of glossary terms that can be used in the project. The list cannot be empty or include over 20 values. Each value must follow the regex of `[a-zA-Z0-9_-]{1,36}$`.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `created_by` - Creator of the project.
* `domain_id` - Id of the project's DataZone domain.
* `id` - Id of the project.
* `name` - Name of the project.
* `created_at` - Timestamp of when the project was made.
* `description` - Description of the project.
* `failure_reasons` - List of error messages if operation cannot be completed.
* `glossary_terms` - Business glossary terms that can be used in the project.
* `last_updated_at` - Timestamp of when the project was last updated.
* `project_status` -  Enum that conveys state of project. Can be `ACTIVE`, `DELETING`, or `DELETE_FAILED`.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `10m`)
* `delete` - (Default `10m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import DataZone Project using a colon-delimited string combining `domain_id` and `id`. For example:

```terraform
import {
  to = aws_datazone_project.example
  id = "domain-1234:project-1234"
}
```

Using `terraform import`, import DataZone Project using a colon-delimited string combining `domain_id` and `id`. For example:

```console
% terraform import aws_datazone_project.example domain-1234:project-1234
```
