---
subcategory: "CodeGuru Reviewer"
layout: "aws"
page_title: "AWS: aws_codegurureviewer_repository_association"
description: |-
  Terraform resource for managing an AWS CodeGuru Reviewer Repository Association.
---

# Resource: aws_codegurureviewer_repository_association

Terraform resource for managing an AWS CodeGuru Reviewer Repository Association.

## Example Usage

```terraform
resource "aws_kms_key" "example" {}

resource "aws_codecommit_repository" "example" {
  repository_name = "example-repo"

  # ignore tag added by CodeGuru service upon association
  lifecycle {
    ignore_changes = [
      tags["codeguru-reviewer"]
    ]
  }
}

resource "aws_codegurureviewer_repository_association" "example" {
  repository {
    codecommit {
      name = aws_codecommit_repository.example.repository_name
    }
  }
  kms_key_details {
    encryption_option = "CUSTOMER_MANAGED_CMK"
    kms_key_id        = aws_kms_key.example.key_id
  }
}

```

## Argument Reference

The following arguments are required:

* `repository` - (Required) An object describing the repository to associate. Valid values: `bitbucket`, `codecommit`, `github_enterprise_server`, or `s3_bucket`. Block is documented below. Note: for repositories that leverage CodeStar connections (ex. `bitbucket`, `github_enterprise_server`) the connection must be in `Available` status prior to creating this resource.

The following arguments are optional:

* `kms_key_details` - (Optional) An object describing the KMS key to asssociate. Block is documented below.

## repository

This configuration block supports the following:

### bitbucket

* `connection_arn` - (Required) The Amazon Resource Name (ARN) of an AWS CodeStar Connections connection.
* `name` - (Required) The name of the third party source repository.
* `owner` - (Required) The username for the account that owns the repository.

### codecommit

* `name` - (Required) The name of the AWS CodeCommit repository.

### github_enterprise_server

* `connection_arn` - (Required) The Amazon Resource Name (ARN) of an AWS CodeStar Connections connection.
* `name` - (Required) The name of the third party source repository.
* `owner` - (Required) The username for the account that owns the repository.

### s3_bucket

* `bucket_name` - (Required) The name of the S3 bucket used for associating a new S3 repository. Note: The name must begin with `codeguru-reviewer-`.
* `name` - (Required) The name of the repository in the S3 bucket.

## kms_key_details

This configuration block supports the following:

* `encryption_option` - (Optional) The encryption option for a repository association. It is either owned by AWS Key Management Service (KMS) (`AWS_OWNED_CMK`) or customer managed (`CUSTOMER_MANAGED_CMK`).
* `kms_key_id` - (Optional) The ID of the AWS KMS key that is associated with a repository association.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - The Amazon Resource Name (ARN) identifying the repository association.
* `association_id` - The ID of the repository association.
* `connection_arn` - The Amazon Resource Name (ARN) of an AWS CodeStar Connections connection.
* `id` - The Amazon Resource Name (ARN) identifying the repository association.
* `name` - The name of the repository.
* `owner` - The owner of the repository.
* `provider_type` - The provider type of the repository association.
* `state` - The state of the repository association.
* `state_reason` - A description of why the repository association is in the current state.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `60m`)
* `update` - (Default `180m`)
* `delete` - (Default `90m`)
