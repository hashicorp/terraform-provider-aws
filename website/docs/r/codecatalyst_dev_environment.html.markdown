---
subcategory: "CodeCatalyst"
layout: "aws"
page_title: "AWS: aws_codecatalyst_dev_environment"
description: |-
  Terraform resource for managing an AWS CodeCatalyst Devenvironment.
---
# Resource: aws_codecatalyst_dev_environment

Terraform resource for managing an AWS CodeCatalyst Dev Environment.

## Example Usage

```terraform
resource "aws_codecatalyst_dev_environment" "test" {
  alias         = "devenv"
  space_name    = "myspace"
  project_name  = "myproject"
  instance_type = "dev.standard1.small"

  persistent_storage {
    size = 16
  }

  ides {
    name    = "PyCharm"
    runtime = "public.ecr.aws/jetbrains/py"
  }

  inactivity_timeout_minutes = 40

  repositories {
    repository_name = "terraform-provider-aws"
    branch_name     = "main"
  }

}
```

## Argument Reference

The following arguments are required:

* `space_name` - (Required) The name of the space.
* `project_name` - (Required) The name of the project in the space.
* `persistent_storage` - (Required) Information about the amount of storage allocated to the Dev Environment.
* `ides` - (Required) Information about the integrated development environment (IDE) configured for a Dev Environment.
* `instance_type` - (Required) The Amazon EC2 instace type to use for the Dev Environment. Valid values include dev.standard1.small,dev.standard1.medium,dev.standard1.large,dev.standard1.xlarge

The following arguments are optional:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `inactivity_timeout_minutes` - (Optional) The amount of time the Dev Environment will run without any activity detected before stopping, in minutes. Only whole integers are allowed. Dev Environments consume compute minutes when running.
* `repositories` - (Optional) The source repository that contains the branch to clone into the Dev Environment.

ides (`ides`) supports the following:

* `name` - (Required) The name of the IDE. Valid values include Cloud9, IntelliJ, PyCharm, GoLand, and VSCode.
* `runtime` - (Required) A link to the IDE runtime image. This parameter is not required if the name is VSCode. Values of the runtime can be for example public.ecr.aws/jetbrains/py,public.ecr.aws/jetbrains/go

repositories (`repositories`) supports the following:

* `repository_name` - (Required) The name of the source repository.
* `branch_name` - (Optional) The name of the branch in a source repository.

persistent storage (` persistent_storage`) supports the following:

* `size` - (Required) The size of the persistent storage in gigabytes (specifically GiB). Valid values for storage are based on memory sizes in 16GB increments. Valid values are 16, 32, and 64.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - Unique identifier for the Dev Environment

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

- `create` - (Default `30m`)
- `update` - (Default `10m`)
- `delete` - (Default `10m`)
