---
subcategory: "CodeArtifact"
layout: "aws"
page_title: "AWS: aws_codeartifact_package_origin_configuration"
description: |-
  Provides a CodeArtifact Package Origin Configuration resource.
---

# Resource: aws_codeartifact_package_origin_configuration

Provides a CodeArtifact Package Origin Configuration resource. Package origin controls determine how new versions of a package can be added to a repository — whether they can be published directly or ingested from an external connection or upstream repository. See [Editing package origin controls](https://docs.aws.amazon.com/codeartifact/latest/ug/package-origin-controls.html) in the AWS CodeArtifact User Guide.

Setting an origin configuration on a package that does not yet exist in the repository creates the package with the given origin controls, which can be used to protect against dependency confusion attacks.

~> **Note:** A package's origin configuration cannot be deleted, only changed. Destroying this resource resets both the `publish` and `upstream` restrictions to `ALLOW` (the CodeArtifact defaults).

## Example Usage

```terraform
resource "aws_codeartifact_domain" "example" {
  domain = "example"
}

resource "aws_codeartifact_repository" "example" {
  repository = "example"
  domain     = aws_codeartifact_domain.example.domain
}

resource "aws_codeartifact_package_origin_configuration" "example" {
  domain     = aws_codeartifact_domain.example.domain
  repository = aws_codeartifact_repository.example.repository
  format     = "npm"
  namespace  = "my-scope"
  package    = "my-internal-package"

  restrictions {
    publish  = "ALLOW"
    upstream = "BLOCK"
  }
}
```

## Argument Reference

This resource supports the following arguments:

* `domain` - (Required) Name of the domain that contains the repository that contains the package.
* `domain_owner` - (Optional) Account number of the AWS account that owns the domain. Defaults to the account of the caller.
* `format` - (Required) Format of the package. Valid values: `npm`, `pypi`, `maven`, `nuget`, `generic`, `ruby`, `swift`, `cargo`.
* `namespace` - (Optional) Namespace of the package. The package component that specifies its namespace depends on its format. For example, the namespace of a Maven package is its `groupId`, and the namespace of an npm package is its scope. Python, NuGet, Ruby, and Cargo packages do not have a namespace.
* `package` - (Required) Name of the package.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#region).
* `repository` - (Required) Name of the repository that contains the package.
* `restrictions` - (Required) Origin restrictions to set on the package. See [`restrictions` Block](#restrictions-block) below.

### `restrictions` Block

The `restrictions` configuration block supports the following arguments:

* `publish` - (Required) Whether new versions of the package can be published directly to the repository. Valid values: `ALLOW`, `BLOCK`.
* `upstream` - (Required) Whether new versions of the package can be added to the repository from an external connection or upstream source. Valid values: `ALLOW`, `BLOCK`.

## Attribute Reference

This resource exports no additional attributes.

## Import

In Terraform v1.12.0 and later, the [`import` block](https://developer.hashicorp.com/terraform/language/import) can be used with the `identity` attribute. For example:

```terraform
import {
  to = aws_codeartifact_package_origin_configuration.example
  identity = {
    domain     = "example"
    repository = "example"
    format     = "npm"
    namespace  = "my-scope"
    package    = "my-internal-package"
  }
}

resource "aws_codeartifact_package_origin_configuration" "example" {
  ### Configuration omitted for brevity ###
}
```

### Identity Schema

#### Required

* `domain` (String) Name of the domain that contains the repository that contains the package.
* `format` (String) Format of the package.
* `package` (String) Name of the package.
* `repository` (String) Name of the repository that contains the package.

#### Optional

* `account_id` (String) AWS Account where this resource is managed.
* `namespace` (String) Namespace of the package.
* `region` (String) Region where this resource is managed.

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import CodeArtifact Package Origin Configurations using the domain, repository, format, namespace, and package separated by commas (the namespace may be empty for formats without one). For example:

```terraform
import {
  to = aws_codeartifact_package_origin_configuration.example
  id = "example,example,npm,my-scope,my-internal-package"
}
```

Using `terraform import`, import CodeArtifact Package Origin Configurations using `domain,repository,format,namespace,package`. For example:

```console
% terraform import aws_codeartifact_package_origin_configuration.example example,example,npm,my-scope,my-internal-package
```
