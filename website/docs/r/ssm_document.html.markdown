---
subcategory: "SSM (Systems Manager)"
layout: "aws"
page_title: "AWS: aws_ssm_document"
description: |-
  Provides an SSM Document resource
---

# Resource: aws_ssm_document

Provides an SSM Document resource

~> **NOTE on updating SSM documents:** Only documents with a schema version of 2.0
or greater can update their content once created, see [SSM Schema Features][1]. To update a document with an older schema version you must recreate the resource. Not all document types support a schema version of 2.0 or greater. Refer to [SSM document schema features and examples][2] for information about which schema versions are supported for the respective `document_type`.

## Example Usage

### Create an ssm document in JSON format

```terraform
resource "aws_ssm_document" "foo" {
  name          = "test_document"
  document_type = "Command"

  content = <<DOC
  {
    "schemaVersion": "1.2",
    "description": "Check ip configuration of a Linux instance.",
    "parameters": {

    },
    "runtimeConfig": {
      "aws:runShellScript": {
        "properties": [
          {
            "id": "0.aws:runShellScript",
            "runCommand": ["ifconfig"]
          }
        ]
      }
    }
  }
DOC
}
```

### Create an ssm document in YAML format

```terraform
resource "aws_ssm_document" "foo" {
  name            = "test_document"
  document_format = "YAML"
  document_type   = "Command"

  content = <<DOC
schemaVersion: '1.2'
description: Check ip configuration of a Linux instance.
parameters: {}
runtimeConfig:
  'aws:runShellScript':
    properties:
      - id: '0.aws:runShellScript'
        runCommand:
          - ifconfig
DOC
}
```

## Argument Reference

This resource supports the following arguments:

* `name` - (Required) The name of the document.
* `attachments_source` - (Optional) One or more configuration blocks describing attachments sources to a version of a document. See [`attachments_source` block](#attachments_source-block) below for details.
* `content` - (Required) The content for the SSM document in JSON or YAML format. The content of the document must not exceed 64KB. This quota also includes the content specified for input parameters at runtime. We recommend storing the contents for your new document in an external JSON or YAML file and referencing the file in a command.
* `document_format` - (Optional, defaults to `JSON`) The format of the document. Valid values: `JSON`, `TEXT`, `YAML`.
* `document_type` - (Required) The type of the document. For a list of valid values, see the [API Reference](https://docs.aws.amazon.com/systems-manager/latest/APIReference/API_CreateDocument.html#systemsmanager-CreateDocument-request-DocumentType).
* `permissions` - (Optional) Additional permissions to attach to the document. See [Permissions](#permissions) below for details.
* `target_type` - (Optional) The target type which defines the kinds of resources the document can run on. For example, `/AWS::EC2::Instance`. For a list of valid resource types, see [AWS resource and property types reference](https://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-template-resource-type-ref.html).
* `tags` - (Optional) A map of tags to assign to the object. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.
* `version_name` - (Optional) The version of the artifact associated with the document. For example, `12.6`. This value is unique across all versions of a document, and can't be changed.

### `attachments_source` block

The `attachments_source` configuration block supports the following arguments:

* `key` - (Required) The key of a key-value pair that identifies the location of an attachment to the document. Valid values: `SourceUrl`, `S3FileUrl`, `AttachmentReference`.
* `values` - (Required) The value of a key-value pair that identifies the location of an attachment to the document. The argument format is a list of a single string that depends on the type of key you specify - see the [API Reference](https://docs.aws.amazon.com/systems-manager/latest/APIReference/API_AttachmentsSource.html) for details.
* `name` - (Optional) The name of the document attachment file.

### Permissions

The `permissions` attribute specifies how you want to share the document. If you share a document privately, you must specify the AWS user account IDs for those people who can use the document. If you share a document publicly, you must specify All as the account ID.

The `permissions` map supports the following:

* `type` - The permission type for the document. The permission type can be `Share`.
* `account_ids` - The AWS user accounts that should have access to the document. The account IDs can either be a group of account IDs or `All`.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - The Amazon Resource Name (ARN) of the document.
* `created_date` - The date the document was created.
* `default_version` - The default version of the document.
* `description` - The description of the document.
* `document_version` - The document version.
* `hash_type` - The hash type of the document. Valid values: `Sha256`, `Sha1`.
* `hash` - The Sha256 or Sha1 hash created by the system when the document was created.
* `id` - The name of the document.
* `latest_version` - The latest version of the document.
* `owner` - The Amazon Web Services user that created the document.
* `parameter` - One or more configuration blocks describing the parameters for the document. See [`parameter` block](#parameter-block) below for details.
* `platform_types` - The list of operating system (OS) platforms compatible with this SSM document. Valid values: `Windows`, `Linux`, `MacOS`.
* `schema_version` - The schema version of the document.
* `status` - The status of the SSM document. Valid values: `Creating`, `Active`, `Updating`, `Deleting`, `Failed`.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

[1]: http://docs.aws.amazon.com/systems-manager/latest/userguide/sysman-ssm-docs.html#document-schemas-features
[2]: https://docs.aws.amazon.com/systems-manager/latest/userguide/document-schemas-features.html

### `parameter` block

The `parameter` configuration block provides the following attributes:

* `default_value` - If specified, the default values for the parameters. Parameters without a default value are required. Parameters with a default value are optional.
* `description` - A description of what the parameter does, how to use it, the default value, and whether or not the parameter is optional.
* `name` - The name of the parameter.
* `type` - The type of parameter. Valid values: `String`, `StringList`.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import SSM Documents using the name. For example:

```terraform
import {
  to = aws_ssm_document.example
  id = "example"
}
```

Using `terraform import`, import SSM Documents using the name. For example:

```console
% terraform import aws_ssm_document.example example
```

The `attachments_source` argument does not have an SSM API method for reading the attachment information detail after creation. If the argument is set in the Terraform configuration on an imported resource, Terraform will always show a difference. To workaround this behavior, either omit the argument from the Terraform configuration or use [`ignore_changes`](https://www.terraform.io/docs/configuration/meta-arguments/lifecycle.html#ignore_changes) to hide the difference. For example:

```terraform
resource "aws_ssm_document" "test" {
  name          = "test_document"
  document_type = "Package"

  attachments_source {
    key    = "SourceUrl"
    values = ["s3://${aws_s3_bucket.object_bucket.bucket}/test.zip"]
  }

  # There is no AWS SSM API for reading attachments_source info directly
  lifecycle {
    ignore_changes = [attachments_source]
  }
}
```
