---
subcategory: "DataZone"
layout: "aws"
page_title: "AWS: aws_datazone_environment_blueprint"
description: |-
  Terraform data source for managing an AWS DataZone Environment Blueprint.
---

# Data Source: aws_datazone_environment_blueprint

Terraform data source for managing an AWS DataZone Environment Blueprint.

## Example Usage

### Basic Usage

```terraform
resource "aws_datazone_domain" "example" {
  name                  = "example_domain"
  domain_execution_role = aws_iam_role.domain_execution_role.arn
}

data "aws_datazone_environment_blueprint" "example" {
  domain_id = aws_datazone_domain.example.id
  name      = "DefaultDataLake"
  managed   = true
}
```

## Argument Reference

This data source supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `domain_id` - (Required) ID of the domain.
* `name` - (Required) Name of the blueprint.
* `managed` (Required) Whether the blueprint is managed by Amazon DataZone.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `id` - ID of the environment blueprint
* `description` - Description of the blueprint
* `blueprint_provider` - Provider of the blueprint
