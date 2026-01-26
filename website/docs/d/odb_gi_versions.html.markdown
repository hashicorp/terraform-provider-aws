---
subcategory: "Oracle Database@AWS"
layout: "AWS: aws_odb_gi_versions_list"
page_title: "AWS: aws_odb_gi_versions"
description: |-
  Terraform data source to retrieve available Grid Infrastructure versions of Oracle Database@AWS.
---

# Data Source: aws_odb_gi_versions

Terraform data source to retrieve available Grid Infrastructure versions of Oracle Database@AWS.

You can find out more about Oracle Database@AWS from [User Guide](https://docs.aws.amazon.com/odb/latest/UserGuide/what-is-odb.html).

## Example Usage

### Basic Usage

```terraform
data "aws_odb_gi_versions" "example" {}

data "aws_odb_gi_versions" "example" {
  shape = "Exadata.X11M"
}

data "aws_odb_gi_versions" "example" {
  shape = "Exadata.X9M"
}
```

## Argument Reference

The following arguments are optional:

* `shape` - (Optional) The system shape.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `gi_versions` - Information about a specific version of Oracle Grid Infrastructure (GI) software that can be installed on a VM cluster.

### gi_versions

* `version` - The GI software version.
