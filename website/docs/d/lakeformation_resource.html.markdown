---
subcategory: "Lake Formation"
layout: "aws"
page_title: "AWS: aws_lakeformation_resource"
description: |-
    Provides details about a Lake Formation resource.
---

# Data Source: aws_lakeformation_resource

Provides details about a Lake Formation resource.

## Example Usage

```terraform
data "aws_lakeformation_resource" "example" {
  arn = "arn:aws:s3:::tf-acc-test-9151654063908211878"
}
```

## Argument Reference

This data source supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `arn` - (Required) ARN of the resource, an S3 path.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `last_modified` - Date and time the resource was last modified in [RFC 3339 format](https://tools.ietf.org/html/rfc3339#section-5.8).
* `role_arn` - Role that the resource was registered with.
