---
subcategory: "OpenSearch"
layout: "aws"
page_title: "AWS: aws_opensearch_package"
description: |-
  Get information on an AWS OpenSearch Package.
---

# Data Source: aws_opensearch_package

Use this data source to get information about an AWS OpenSearch Package.

## Example Usage

### Basic Usage

```terraform
data "aws_opensearch_package" "example" {
  package_name = "example-txt"
}
```

### Cross-State Reference with Package Association

Look up a package managed in another Terraform state and associate it with a domain.

```terraform
data "aws_opensearch_package" "example" {
  package_name = "example-txt"
}

resource "aws_opensearch_package_association" "example" {
  package_id  = data.aws_opensearch_package.example.package_id
  domain_name = aws_opensearch_domain.example.domain_name
}
```

## Argument Reference

This data source supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `package_name` - (Required) Name of the package to look up.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `id` - ID of the package.
* `available_package_version` - Current version of the package.
* `engine_version` - Engine version that the package is compatible with.
* `package_description` - Description of the package.
* `package_id` - ID of the package.
* `package_type` - Type of the package. Valid values are `TXT-DICTIONARY`, `ZIP-PLUGIN`, `PACKAGE-LICENSE` and `PACKAGE-CONFIG`.
