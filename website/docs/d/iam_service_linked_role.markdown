---
subcategory: "IAM (Identity & Access Management)"
layout: "aws"
page_title: "AWS: aws_iam_service_linked_role"
description: |-
  Get information on a Amazon IAM Service Linked role
---

# Data Source: aws_iam_service_linked_role

This data source can be used to fetch information about a specific
IAM Service Linked role. By using this data source, you can reference IAM role
properties without having to hard code ARNs as input.

~> **NOTE:** This data source can ensure if the Service Linked role exists. In this scenario, if `create_if_missing` is set, Terraform will not _create_ or _adopt_ this resource, but instead will create the role to ensure if exists before exporting the data. Please use the `aws_iam_service_linked_role` resource to manage a service linked role.

## Example Usage

```terraform
data "aws_iam_service_linked_role" "example" {
  aws_service_name = "elasticbeanstalk.amazonaws.com"
}
```

## Argument Reference

This data source supports the following arguments:

* `aws_service_name` - (Required) The AWS service to which this role is attached. You use a string similar to a URL but without the `http://` in front. For example: `elasticbeanstalk.amazonaws.com`. To find the full list of services that support service-linked roles, check [the docs](https://docs.aws.amazon.com/IAM/latest/UserGuide/reference_aws-services-that-work-with-iam.html).
* `custom_suffix` - (Optional) Additional string appended to the role name. Not all AWS services support custom suffixes.
* `create_if_missing` - (Optional) This will create the service linked role if it does not exists. Default value is `false`

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `id` - The Amazon Resource Name (ARN) of the role.
* `arn` - The Amazon Resource Name (ARN) specifying the role.
* `create_date` - The creation date of the IAM role.
* `name` - The name of the role.
* `path` - The path of the role.
* `unique_id` - The stable and unique string identifying the role.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).
