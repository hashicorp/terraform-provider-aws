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

## Argument Reference

The following arguments are supported:

* `description` - (Optional) A textual description for the workflow.
* `on_exception_steps` - (Optional) Specifies the steps (actions) to take if errors are encountered during execution of the workflow. See Workflow Steps below.
* `steps` - (Required) Specifies the details for the steps that are in the specified workflow. See Workflow Steps below.
* `tags` - (Optional) A map of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](https://www.terraform.io/docs/providers/aws/index.html#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

### Workflow Steps

* `copy_step_details` - (Optional) Details for a step that performs a file copy. See Copy Step Details below.
* `custom_step_details` - (Optional) Details for a step that invokes a lambda function.
* `delete_step_details` - (Optional) Details for a step that deletes the file.
* `tag_step_details` - (Optional) Details for a step that creates one or more tags.
* `type` - (Required) One of the following step types are supported. `COPY`, `CUSTOM`, `DELETE`, and `TAG`.

#### Copy Step Details

* `destination_file_location` - (Optional) Specifies the location for the file being copied. Use ${Transfer:username} in this field to parametrize the destination prefix by username.
* `name` - (Optional) The name of the step, used as an identifier.
* `overwrite_existing` - (Optional) A flag that indicates whether or not to overwrite an existing file of the same name. The default is `FALSE`. Valid values are `TRUE` and `FALSE`.
* `source_file_location` - (Optional) Specifies which file to use as input to the workflow step: either the output from the previous step, or the originally uploaded file for the workflow. Enter ${previous.file} to use the previous file as the input. In this case, this workflow step uses the output file from the previous workflow step as input. This is the default value. Enter ${original.file} to use the originally-uploaded file location as input for this step.

#### Custom Step Details

* `name` - (Optional) The name of the step, used as an identifier.
* `source_file_location` - (Optional) Specifies which file to use as input to the workflow step: either the output from the previous step, or the originally uploaded file for the workflow. Enter ${previous.file} to use the previous file as the input. In this case, this workflow step uses the output file from the previous workflow step as input. This is the default value. Enter ${original.file} to use the originally-uploaded file location as input for this step.
* `target` - (Optional) The ARN for the lambda function that is being called.
* `timeout_seconds` - (Optional) Timeout, in seconds, for the step.

#### Delete Step Details

* `name` - (Optional) The name of the step, used as an identifier.
* `source_file_location` - (Optional) Specifies which file to use as input to the workflow step: either the output from the previous step, or the originally uploaded file for the workflow. Enter ${previous.file} to use the previous file as the input. In this case, this workflow step uses the output file from the previous workflow step as input. This is the default value. Enter ${original.file} to use the originally-uploaded file location as input for this step.

#### Tag Step Details

* `name` - (Optional) The name of the step, used as an identifier.
* `source_file_location` - (Optional) Specifies which file to use as input to the workflow step: either the output from the previous step, or the originally uploaded file for the workflow. Enter ${previous.file} to use the previous file as the input. In this case, this workflow step uses the output file from the previous workflow step as input. This is the default value. Enter ${original.file} to use the originally-uploaded file location as input for this step.
* `tags` - (Optional) Array that contains from 1 to 10 key/value pairs. See S3 Tags below.

##### Destination File Location

* `efs_file_location` - (Optional) Specifies the details for the EFS file being copied.
* `s3_file_location` - (Optional) Specifies the details for the S3 file being copied.

###### EFS File Location

* `file_system_id` - (Optional) The ID of the file system, assigned by Amazon EFS.
* `path` - (Optional) The pathname for the folder being used by a workflow.

###### S3 File Location

* `bucket` - (Optional) Specifies the S3 bucket for the customer input file.
* `key` - (Optional) The name assigned to the file when it was created in S3. You use the object key to retrieve the object.

##### S3 Tag

* `key` - (Required) The name assigned to the tag that you create.
* `value` - (Required) The value that corresponds to the key.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - The Workflow ARN.
* `id` - The Workflow id.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://www.terraform.io/docs/providers/aws/index.html#default_tags-configuration-block).

## Import

Transfer Workflows can be imported using the `worflow_id`.

```
$ terraform import aws_transfer_workflow.example example
```
