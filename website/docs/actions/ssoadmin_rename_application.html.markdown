---
subcategory: "SSO Admin"
layout: "aws"
page_title: "AWS: aws_ssoadmin_rename_application"
description: |-
  Renames an AWS SSO Admin application.
---

# Action: aws_ssoadmin_rename_application

~> **Note:** `aws_ssoadmin_rename_application` is in beta. Its interface and behavior may change as the feature evolves, and breaking changes are possible. It is offered as a technical preview without compatibility guarantees until Terraform 1.14 is generally available.

Renames an AWS SSO Admin application. This action allows for imperative renaming of SSO applications.

For information about AWS SSO Admin applications, see the [AWS SSO Admin User Guide](https://docs.aws.amazon.com/singlesignon/latest/userguide/). For specific information about updating applications, see the [UpdateApplication](https://docs.aws.amazon.com/singlesignon/latest/APIReference/API_UpdateApplication.html) page in the AWS SSO Admin API Reference.

## Example Usage

### Basic Usage

```terraform
resource "aws_ssoadmin_application" "example" {
  name                     = "example-app"
  description              = "Example SSO application"
  instance_arn             = tolist(data.aws_ssoadmin_instances.example.arns)[0]
  application_provider_arn = data.aws_ssoadmin_application_providers.example.application_providers[0].application_provider_arn
}

action "aws_ssoadmin_rename_application" "example" {
  config {
    application_arn = aws_ssoadmin_application.example.application_arn
    new_name        = "renamed-example-app"
  }
}

resource "terraform_data" "example" {
  input = "rename-application"

  lifecycle {
    action_trigger {
      events  = [before_update]
      actions = [action.aws_ssoadmin_rename_application.example]
    }
  }
}
```

### Conditional Renaming

```terraform
action "aws_ssoadmin_rename_application" "conditional" {
  config {
    application_arn = aws_ssoadmin_application.example.application_arn
    new_name        = var.environment == "production" ? "prod-${var.app_name}" : "dev-${var.app_name}"
  }
}
```

### Batch Renaming with Prefix

```terraform
action "aws_ssoadmin_rename_application" "batch_rename" {
  config {
    application_arn = aws_ssoadmin_application.example.application_arn
    new_name        = "${var.naming_prefix}-${aws_ssoadmin_application.example.name}"
  }
}
```

## Argument Reference

This action supports the following arguments:

* `application_arn` - (Required) The ARN of the SSO application to rename. This uniquely identifies the application within your AWS account.
* `new_name` - (Required) The new name for the SSO application. The name must be unique within the SSO instance and follow AWS naming conventions.
* `region` - (Optional) Region where this action should be [run](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
