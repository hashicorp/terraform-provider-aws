---
subcategory: "Oracle Database@AWS"
layout: "AWS: aws_odb_cloud_exadata_infrastructures"
page_title: "AWS: aws_odb_cloud_exadata_infrastructures"
description: |-
  Terraform data source for managing exadata infrastructures in AWS for Oracle Database@AWS.
---

# Data Source: aws_odb_cloud_exadata_infrastructures

Terraform data source for exadata infrastructures in AWS for Oracle Database@AWS.

You can find out more about Oracle Database@AWS from [User Guide](https://docs.aws.amazon.com/odb/latest/UserGuide/what-is-odb.html).

## Example Usage

### Basic Usage

```terraform
data "aws_odb_cloud_exadata_infrastructures" "example" {}
```

## Argument Reference

The following arguments are optional:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `cloud_exadata_infrastructures` - List of Cloud Exadata Infrastructures. Returns basic information about the Cloud Exadata Infrastructures.

### cloud_exadata_infrastructures

* `arn` - The Amazon Resource Name (ARN) for the Exadata infrastructure.
* `id`  - The unique identifier of the Exadata infrastructure.
* `oci_resource_anchor_name` - The name of the OCI resource anchor for the Exadata infrastructure.
* `oci_url` - The HTTPS link to the Exadata infrastructure in OCI.
* `ocid` - The OCID of the Exadata infrastructure in OCI.
* `display_name` - The display name of the Exadata infrastructure.
