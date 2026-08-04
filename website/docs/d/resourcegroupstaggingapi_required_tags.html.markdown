---
subcategory: "Resource Groups Tagging"
layout: "aws"
page_title: "AWS: aws_resourcegroupstaggingapi_required_tags"
description: |-
  Lists required tags for supported resource types in an AWS account.
---

# Data Source: aws_resourcegroupstaggingapi_required_tags

Lists the required tags for supported resource types in an AWS account. Required tags are defined through AWS Organizations [tag policies](https://docs.aws.amazon.com/organizations/latest/userguide/orgs_manage_policies_tag-policies.html).

## Example Usage

### Basic Usage

```terraform
data "aws_resourcegroupstaggingapi_required_tags" "example" {}
```

## Argument Reference

This data source supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `required_tags` - List of required tag configurations. See [`required_tags`](#required_tags) below.

### `required_tags`

* `cloud_formation_resource_types` - CloudFormation resource types assigned the required tag keys.
* `reporting_tag_keys` - Tag keys marked as required in the `report_required_tag_for` block of the effective tag policy.
* `resource_type` - Resource type for the required tag keys.
