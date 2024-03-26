---
subcategory: "EC2 Image Builder"
layout: "aws"
page_title: "AWS: aws_imagebuilder_workflow"
description: |-
  Terraform resource for managing an AWS EC2 Image Builder Workflow.
---
# Resource: aws_imagebuilder_workflow

Terraform resource for managing an AWS EC2 Image Builder Workflow.

## Example Usage

### Basic Usage

```terraform
resource "aws_imagebuilder_workflow" "example" {
  name    = "example"
  version = "1.0.0"
  type    = "TEST"

  data = <<-EOT
  name: example
  description: Workflow to test an image
  schemaVersion: 1.0

  parameters:
    - name: waitForActionAtEnd
      type: boolean

  steps:
    - name: LaunchTestInstance
      action: LaunchInstance
      onFailure: Abort
      inputs:
        waitFor: "ssmAgent"

    - name: TerminateTestInstance
      action: TerminateInstance
      onFailure: Continue
      inputs:
        instanceId.$: "$.stepOutputs.LaunchTestInstance.instanceId"

    - name: WaitForActionAtEnd
      action: WaitForAction
      if:
        booleanEquals: true
        value: "$.parameters.waitForActionAtEnd"
  EOT
}
```

## Argument Reference

The following arguments are required:

* `name` - (Required) Name of the workflow.
* `type` - (Required) Type of the workflow. Valid values: `BUILD`, `TEST`, `DISTRIBUTION`.
* `version` - (Required) Version of the workflow.

The following arguments are optional:

* `change_description` - (Optional) Change description of the workflow.
* `data` - (Optional) Inline YAML string with data of the workflow. Exactly one of `data` and `uri` can be specified.
* `description` - (Optional) Description of the workflow.
* `kms_key_id` - (Optional) Amazon Resource Name (ARN) of the Key Management Service (KMS) Key used to encrypt the workflow.
* `tags` - (Optional) Key-value map of resource tags for the workflow. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.
* `uri` - (Optional) S3 URI with data of the workflow. Exactly one of `data` and `uri` can be specified.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - Amazon Resource Name (ARN) of the workflow.
* `date_created` - Date the workflow was created.
* `owner` - Owner of the workflow.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import EC2 Image Builder Workflow using the `example_id_arg`. For example:

```terraform
import {
  to = aws_imagebuilder_workflow.example
  id = "workflow-id-12345678"
}
```

Using `terraform import`, import EC2 Image Builder Workflow using the `example_id_arg`. For example:

```console
% terraform import aws_imagebuilder_workflow.example arn:aws:imagebuilder:us-east-1:aws:workflow/test/example/1.0.1/1
```

Certain resource arguments, such as `uri`, cannot be read via the API and imported into Terraform. Terraform will display a difference for these arguments the first run after import if declared in the Terraform configuration for an imported resource.
