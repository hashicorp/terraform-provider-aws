---
subcategory: "ECR Public"
layout: "aws"
page_title: "AWS: aws_ecrpublic_repository"
description: |-
  Provides a Public Elastic Container Registry Repository.
---

# Resource: aws_ecrpublic_repository

Provides a Public Elastic Container Registry Repository.

~> **NOTE:** This resource can only be used in the `us-east-1` region.

## Example Usage

```terraform
provider "aws" {
  alias  = "us_east_1"
  region = "us-east-1"
}

resource "aws_ecrpublic_repository" "foo" {
  provider = aws.us_east_1

  repository_name = "bar"

  catalog_data {
    about_text        = "About Text"
    architectures     = ["ARM"]
    description       = "Description"
    logo_image_blob   = filebase64(image.png)
    operating_systems = ["Linux"]
    usage_text        = "Usage Text"
  }

  tags = {
    env = "production"
  }
}
```

## Argument Reference

This resource supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `repository_name` - (Required) Name of the repository.
* `catalog_data` - (Optional) Catalog data configuration for the repository. See [below for schema](#catalog_data).
* `tags` - (Optional) Key-value mapping of resource tags. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

### catalog_data

* `about_text` - (Optional) A detailed description of the contents of the repository. It is publicly visible in the Amazon ECR Public Gallery. The text must be in markdown format.
* `architectures` - (Optional) The system architecture that the images in the repository are compatible with. On the Amazon ECR Public Gallery, the following supported architectures will appear as badges on the repository and are used as search filters: `ARM`, `ARM 64`, `x86`, `x86-64`
* `description` - (Optional) A short description of the contents of the repository. This text appears in both the image details and also when searching for repositories on the Amazon ECR Public Gallery.
* `logo_image_blob` - (Optional) The base64-encoded repository logo payload. (Only visible for verified accounts) Note that drift detection is disabled for this attribute.
* `operating_systems` -  (Optional) The operating systems that the images in the repository are compatible with. On the Amazon ECR Public Gallery, the following supported operating systems will appear as badges on the repository and are used as search filters: `Linux`, `Windows`
* `usage_text` -  (Optional) Detailed information on how to use the contents of the repository. It is publicly visible in the Amazon ECR Public Gallery. The usage text provides context, support information, and additional usage details for users of the repository. The text must be in markdown format.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - Full ARN of the repository.
* `id` - The repository name.
* `registry_id` - The registry ID where the repository was created.
* `repository_uri` - The URI of the repository.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

- `delete` - (Default `20m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import ECR Public Repositories using the `repository_name`. For example:

```terraform
import {
  to = aws_ecrpublic_repository.example
  id = "example"
}
```

Using `terraform import`, import ECR Public Repositories using the `repository_name`. For example:

```console
% terraform import aws_ecrpublic_repository.example example
```
