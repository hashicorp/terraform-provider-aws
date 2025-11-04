---
subcategory: "Transfer Family"
layout: "aws"
page_title: "AWS: aws_transfer_web_app_customization"
description: |-
  Terraform resource for managing an AWS Transfer Family Web App Customization.
---

# Resource: aws_transfer_web_app_customization

Terraform resource for managing an AWS Transfer Family Web App Customization.

## Example Usage

### Basic Usage

```terraform
resource "aws_transfer_web_app" "test" {
  identity_provider_details {
    identity_center_config {
      instance_arn = tolist(data.aws_ssoadmin_instances.test.arns)[0]
      role         = aws_iam_role.test.arn
    }
  }
  web_app_units {
    provisioned = 1
  }
  tags = {
    Name = "test"
  }
}

resource "aws_transfer_web_app_customization" "test" {
  web_app_id   = aws_transfer_web_app.test.web_app_id
  favicon_file = filebase64("${path.module}/favicon.png")
  logo_file    = filebase64("${path.module}/logo.png")
  title        = "test"
}
```

## Argument Reference

The following arguments are required:

* `web_app_id` - (Required) The identifier of the web app to be customized.

The following arguments are optional:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `favicon_file` - (Optional) Base64-encoded string representing the favicon image. Terraform will detect drift only if this argument is specified. To remove the favicon, recreate the resource.
* `logo_file` - (Optional) Base64-encoded string representing the logo image. Terraform will detect drift only if this argument is specified. To remove the logo, recreate the resource.
* `title` – (Optional) Title of the web app. Must be between 1 and 100 characters in length (an empty string is not allowed). To remove the title, omit this argument from your configuration.

## Attribute Reference

This resource exports no additional attributes.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Transfer Family Web App Customization using the `web_app_id`. For example:

```terraform
import {
  to = aws_transfer_web_app_customization.example
  id = "webapp-12345678901234567890"
}
```

Using `terraform import`, import Transfer Family Web App Customization using the `web_app_id`. For example:

```console
% terraform import aws_transfer_web_app_customization.example webapp-12345678901234567890 
```
