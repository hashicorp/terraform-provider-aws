---
subcategory: "Transfer Family"
layout: "aws"
page_title: "AWS: aws_transfer_workflow"
description: |-
  Provides a AWS Transfer Workflow resource.
---

# Resource: aws_transfer_workflow

Provides a AWS Transfer Workflow resource.

## Example Usage

### Basic single step example

```terraform
resource "aws_transfer_workflow" "example" {
  steps {
    delete_step_details {
      name                 = "example"
      source_file_location = "$${original.file}"
    }
    type = "DELETE"
  }
}
```

### Multistep example

```terraform
resource "aws_transfer_workflow" "example" {
  steps {
    custom_step_details {
      name                 = "example"
      source_file_location = "$${original.file}"
      target               = aws_lambda_function.example.arn
      timeout_seconds      = 60
    }
    type = "CUSTOM"
  }

  steps {
    tag_step_details {
      name                 = "example"
      source_file_location = "$${original.file}"
      tags {
        key   = "Name"
        value = "Hello World"
      }
    }
    type = "TAG"
  }
}
```

## Argument Reference

This resource supports the following arguments:

* `description` - (Optional) Textual description for the workflow.
* `on_exception_steps` - (Optional) Steps (actions) to take if errors are encountered during execution of the workflow. See [`on_exception_steps` Block](#on_exception_steps-block) below.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `steps` - (Required) Details for the steps that are in the specified workflow. See [`steps` Block](#steps-block) below.
* `tags` - (Optional) Map of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

### `on_exception_steps` Block

* `copy_step_details` - (Optional) Details for a step that performs a file copy. See [`copy_step_details` Block](#copy_step_details-block) below.
* `custom_step_details` - (Optional) Details for a step that invokes a lambda function. See [`custom_step_details` Block](#custom_step_details-block) below.
* `decrypt_step_details` - (Optional) Details for a step that decrypts the file. See [`decrypt_step_details` Block](#decrypt_step_details-block) below.
* `delete_step_details` - (Optional) Details for a step that deletes the file. See [`delete_step_details` Block](#delete_step_details-block) below.
* `tag_step_details` - (Optional) Details for a step that creates one or more tags. See [`tag_step_details` Block](#tag_step_details-block) below.
* `type` - (Required) Step type. Valid values are `COPY`, `CUSTOM`, `DECRYPT`, `DELETE`, and `TAG`.

### `steps` Block

* `copy_step_details` - (Optional) Details for a step that performs a file copy. See [`copy_step_details` Block](#copy_step_details-block) below.
* `custom_step_details` - (Optional) Details for a step that invokes a lambda function. See [`custom_step_details` Block](#custom_step_details-block) below.
* `decrypt_step_details` - (Optional) Details for a step that decrypts the file. See [`decrypt_step_details` Block](#decrypt_step_details-block) below.
* `delete_step_details` - (Optional) Details for a step that deletes the file. See [`delete_step_details` Block](#delete_step_details-block) below.
* `tag_step_details` - (Optional) Details for a step that creates one or more tags. See [`tag_step_details` Block](#tag_step_details-block) below.
* `type` - (Required) Step type. Valid values are `COPY`, `CUSTOM`, `DECRYPT`, `DELETE`, and `TAG`.

### `copy_step_details` Block

* `destination_file_location` - (Optional) Location for the file being copied. Use `${Transfer:username}` in this field to parametrize the destination prefix by username. See [`destination_file_location` Block](#destination_file_location-block) below.
* `name` - (Optional) Name of the step, used as an identifier.
* `overwrite_existing` - (Optional) Flag that indicates whether or not to overwrite an existing file of the same name. The default is `FALSE`. Valid values are `TRUE` and `FALSE`.
* `source_file_location` - (Optional) File to use as input to the workflow step: either the output from the previous step, or the originally uploaded file for the workflow. Enter `${previous.file}` to use the previous file as the input. In this case, this workflow step uses the output file from the previous workflow step as input. This is the default value. Enter `${original.file}` to use the originally-uploaded file location as input for this step.

### `custom_step_details` Block

* `name` - (Optional) Name of the step, used as an identifier.
* `source_file_location` - (Optional) File to use as input to the workflow step: either the output from the previous step, or the originally uploaded file for the workflow. Enter `${previous.file}` to use the previous file as the input. In this case, this workflow step uses the output file from the previous workflow step as input. This is the default value. Enter `${original.file}` to use the originally-uploaded file location as input for this step.
* `target` - (Optional) ARN for the lambda function that is being called.
* `timeout_seconds` - (Optional) Timeout, in seconds, for the step.

### `decrypt_step_details` Block

* `destination_file_location` - (Optional) Location for the file being copied. Use `${Transfer:username}` in this field to parametrize the destination prefix by username. See [`destination_file_location` Block](#destination_file_location-block) below.
* `name` - (Optional) Name of the step, used as an identifier.
* `overwrite_existing` - (Optional) Flag that indicates whether or not to overwrite an existing file of the same name. The default is `FALSE`. Valid values are `TRUE` and `FALSE`.
* `source_file_location` - (Optional) File to use as input to the workflow step: either the output from the previous step, or the originally uploaded file for the workflow. Enter `${previous.file}` to use the previous file as the input. In this case, this workflow step uses the output file from the previous workflow step as input. This is the default value. Enter `${original.file}` to use the originally-uploaded file location as input for this step.
* `type` - (Required) Type of encryption used. Currently, this value must be `"PGP"`.

### `delete_step_details` Block

* `name` - (Optional) Name of the step, used as an identifier.
* `source_file_location` - (Optional) File to use as input to the workflow step: either the output from the previous step, or the originally uploaded file for the workflow. Enter `${previous.file}` to use the previous file as the input. In this case, this workflow step uses the output file from the previous workflow step as input. This is the default value. Enter `${original.file}` to use the originally-uploaded file location as input for this step.

### `tag_step_details` Block

* `name` - (Optional) Name of the step, used as an identifier.
* `source_file_location` - (Optional) File to use as input to the workflow step: either the output from the previous step, or the originally uploaded file for the workflow. Enter `${previous.file}` to use the previous file as the input. In this case, this workflow step uses the output file from the previous workflow step as input. This is the default value. Enter `${original.file}` to use the originally-uploaded file location as input for this step.
* `tags` - (Optional) Array that contains from 1 to 10 key/value pairs. See [`tags` Block](#tags-block) below.

### `destination_file_location` Block

* `efs_file_location` - (Optional) Details for the EFS file being copied. See [`efs_file_location` Block](#efs_file_location-block) below.
* `s3_file_location` - (Optional) Details for the S3 file being copied. See [`s3_file_location` Block](#s3_file_location-block) below.

### `efs_file_location` Block

* `file_system_id` - (Optional) ID of the file system, assigned by Amazon EFS.
* `path` - (Optional) Pathname for the folder being used by a workflow.

### `s3_file_location` Block

* `bucket` - (Optional) S3 bucket for the customer input file.
* `key` - (Optional) Name assigned to the file when it was created in S3. You use the object key to retrieve the object.

### `tags` Block

* `key` - (Required) Name assigned to the tag that you create.
* `value` - (Required) Value that corresponds to the key.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - Workflow ARN.
* `id` - Workflow ID.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Transfer Workflows using the `worflow_id`. For example:

```terraform
import {
  to = aws_transfer_workflow.example
  id = "example"
}
```

Using `terraform import`, import Transfer Workflows using the `worflow_id`. For example:

```console
% terraform import aws_transfer_workflow.example example
```
