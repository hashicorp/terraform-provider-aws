---
subcategory: "SSM (Systems Manager)"
layout: "aws"
page_title: "AWS: aws_ssm_parameters_by_path"
description: |-
  Provides SSM Parameters by path
---

# Data Source: aws_ssm_parameters_by_path

Use this data source to get information about one or more System Manager parameters in a specific hierarchy.

## Example Usage

```terraform
data "aws_ssm_parameters_by_path" "example" {
  path = "/site/newyork/department/" # Trailing slash is optional
}
```

~> **Note:** When the `with_decryption` argument is set to `true`, the unencrypted values of `SecureString` parameters will be stored in the raw state as plain-text as per normal Terraform behavior. [Read more about sensitive data in state](/docs/state/sensitive-data.html).

~> **Note:** The data source follows the behavior of the [SSM API](https://docs.aws.amazon.com/sdk-for-go/api/service/ssm/#Parameter) to return a string value, regardless of parameter type. For `StringList` type where the value is returned as a comma-separated string with no spaces between comma, you may use the built-in [split](https://www.terraform.io/docs/configuration/functions/split.html) function to get values in a list. Example: `split(",", data.aws_ssm_parameter.subnets.value)`

## Argument Reference

This data source supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `path` - (Required) The hierarchy for the parameter. Hierarchies start with a forward slash (/). The hierarchy is the parameter name except the last part of the parameter. The last part of the parameter name can't be in the path. A parameter name hierarchy can have a maximum of 15 levels. **Note:** If the parameter name (e.g., `/my-app/my-param`) is specified, the data source will not retrieve any value as designed, unless there are other parameters that happen to use the former path in their hierarchy (e.g., `/my-app/my-param/my-actual-param`).
* `with_decryption` - (Optional) Whether to retrieve all parameters in the hierarchy, particularly those of `SecureString` type, with their value decrypted. Defaults to `true`.
* `recursive` - (Optional) Whether to retrieve all parameters within the hirerachy. Defaults to `false`.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `arns` - A list that contains the Amazon Resource Names (ARNs) of the retrieved parameters.
* `names` - A list that contains the names of the retrieved parameters.
* `types` - A list that contains the types (`String`, `StringList`, or `SecureString`) of retrieved parameters.
* `values` - A list that contains the retrieved parameter values. **Note:** This value is always marked as sensitive in the Terraform plan output, regardless of whether any retrieved parameters are of `SecureString` type. Use the [`nonsensitive` function](https://developer.hashicorp.com/terraform/language/functions/nonsensitive) to override the behavior at your own risk and discretion, if you are certain that there are no sensitive values being retrieved.
