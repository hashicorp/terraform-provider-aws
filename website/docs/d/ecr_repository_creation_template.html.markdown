---
subcategory: "ECR (Elastic Container Registry)"
layout: "aws"
page_title: "AWS: aws_ecr_repository_creation_template"
description: |-
  Provides details about an ECR Repository Creation Template
---

# Data Source: aws_ecr_repository_creation_template

The ECR Repository Creation Template data source allows the template details to be retrieved for a Repository Creation Template.

## Example Usage

```terraform
data "aws_ecr_repository_creation_template" "example" {
  prefix = "example"
}
```

## Argument Reference

This data source supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `prefix` - (Required) The repository name prefix that the template matches against.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `applied_for` - Which features this template applies to. Contains one or more of `PULL_THROUGH_CACHE` or `REPLICATION`.
* `custom_role_arn` - The ARN of the custom role used for repository creation.
* `description` - The description for this template.
* `encryption_configuration` - Encryption configuration for any created repositories. See [Encryption Configuration](#encryption-configuration) below.
* `image_tag_mutability` - The tag mutability setting for any created repositories.
* `lifecycle_policy` - The lifecycle policy document to apply to any created repositories.
* `registry_id` - The registry ID the repository creation template applies to.
* `repository_policy` - The registry policy document to apply to any created repositories.
* `resource_tags` - A map of tags to assign to any created repositories.

### Encryption Configuration

* `encryption_type` - Encryption type to use for any created repositories, either `AES256` or `KMS`.
* `kms_key` - If `encryption_type` is `KMS`, the ARN of the KMS key used.
