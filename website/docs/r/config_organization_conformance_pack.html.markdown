---
subcategory: "Config"
layout: "aws"
page_title: "AWS: aws_config_organization_conformance_pack"
description: |-
  Manages a Config Organization Conformance Pack
---

# Resource: aws_config_organization_conformance_pack

Manages a Config Organization Conformance Pack. More information can be found in the [Managing Conformance Packs Across all Accounts in Your Organization](https://docs.aws.amazon.com/config/latest/developerguide/conformance-pack-organization-apis.html) and [AWS Config Managed Rules](https://docs.aws.amazon.com/config/latest/developerguide/evaluate-config_use-managed-rules.html) documentation. Example conformance pack templates may be found in the [AWS Config Rules Repository](https://github.com/awslabs/aws-config-rules/tree/master/aws-config-conformance-packs).

~> **NOTE:** This resource must be created in the Organization master account or a delegated administrator account, and the Organization must have all features enabled. Every Organization account except those configured in the `excluded_accounts` argument must have a Configuration Recorder with proper IAM permissions before the Organization Conformance Pack will successfully create or update. See also the [`aws_config_configuration_recorder` resource](/docs/providers/aws/r/config_configuration_recorder.html).

## Example Usage

### Using Template Body

```hcl
resource "aws_config_organization_conformance_pack" "example" {
  name = "example"

  input_parameter {
    parameter_name  = "AccessKeysRotatedParameterMaxAccessKeyAge"
    parameter_value = "90"
  }

  template_body = <<EOT
Parameters:
  AccessKeysRotatedParameterMaxAccessKeyAge:
    Type: String
Resources:
  IAMPasswordPolicy:
    Properties:
      ConfigRuleName: IAMPasswordPolicy
      Source:
        Owner: AWS
        SourceIdentifier: IAM_PASSWORD_POLICY
    Type: AWS::Config::ConfigRule
EOT

  depends_on = [aws_config_configuration_recorder.example, aws_organizations_organization.example]
}

resource "aws_organizations_organization" "example" {
  aws_service_access_principals = ["config-multiaccountsetup.amazonaws.com"]
  feature_set                   = "ALL"
}
```

### Using Template S3 URI

```hcl
resource "aws_config_organization_conformance_pack" "example" {
  name            = "example"
  template_s3_uri = "s3://${aws_s3_bucket.example.bucket}/${aws_s3_bucket_object.example.key}"

  depends_on = [aws_config_configuration_recorder.example, aws_organizations_organization.example]
}

resource "aws_organizations_organization" "example" {
  aws_service_access_principals = ["config-multiaccountsetup.amazonaws.com"]
  feature_set                   = "ALL"
}

resource "aws_s3_bucket" "example" {
  bucket = "example"
}

resource "aws_s3_bucket_object" "example" {
  bucket  = aws_s3_bucket.example.id
  key     = "example-key"
  content = <<EOT
Resources:
  IAMPasswordPolicy:
    Properties:
      ConfigRuleName: IAMPasswordPolicy
      Source:
        Owner: AWS
        SourceIdentifier: IAM_PASSWORD_POLICY
    Type: AWS::Config::ConfigRule
EOT
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required, Forces new resource) The name of the organization conformance pack. Must begin with a letter and contain from 1 to 128 alphanumeric characters and hyphens.
* `delivery_s3_bucket` - (Optional) Amazon S3 bucket where AWS Config stores conformance pack templates. Delivery bucket must begin with `awsconfigconforms` prefix. Maximum length of 63.
* `delivery_s3_key_prefix` - (Optional) The prefix for the Amazon S3 bucket. Maximum length of 1024.
* `excluded_accounts` - (Optional) Set of AWS accounts to be excluded from an organization conformance pack while deploying a conformance pack. Maximum of 1000 accounts.
* `input_parameter` - (Optional) Set of configuration blocks describing input parameters passed to the conformance pack template. Documented below. When configured, the parameters must also be included in the `template_body` or in the template stored in Amazon S3 if using `template_s3_uri`.
* `template_body` - (Optional, Conflicts with `template_s3_uri`) A string containing full conformance pack template body. Maximum length of 51200. Drift detection is not possible with this argument.
* `template_s3_uri` - (Optional, Conflicts with `template_body`) Location of file, e.g., `s3://bucketname/prefix`, containing the template body. The uri must point to the conformance pack template that is located in an Amazon S3 bucket in the same region as the conformance pack. Maximum length of 1024. Drift detection is not possible with this argument.

### input_parameter Argument Reference

The `input_parameter` configuration block supports the following arguments:

* `parameter_name` - (Required) The input key.
* `parameter_value` - (Required) The input value.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - Amazon Resource Name (ARN) of the organization conformance pack.
* `id` - The name of the organization conformance pack.

## Timeouts

`aws_config_organization_conformance_pack` provides the following
[Timeouts](https://www.terraform.io/docs/configuration/blocks/resources/syntax.html#operation-timeouts) configuration options:

- `create` - (Default `10 minutes`) Used for creating conformance pack
- `update` - (Default `10 minutes`) Used for conformance pack modifications
- `delete` - (Default `20 minutes`) Used for destroying conformance pack

## Import

Config Organization Conformance Packs can be imported using the `name`, e.g.,

```
$ terraform import aws_config_organization_conformance_pack.example example
```
