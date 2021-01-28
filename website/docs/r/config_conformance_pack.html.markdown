---
subcategory: "Config"
layout: "aws"
page_title: "AWS: aws_config_conformance_pack"
description: |-
  Manages a Config Conformance Pack
---

# Resource: aws_config_conformance_pack

Manages a Config Conformance Pack. More information about these rules can be found in the
[Conformance Packs](https://docs.aws.amazon.com/config/latest/developerguide/conformance-packs.html) documentation.
Example conformance pack templates may be found in the
[AWS Config Rules Repository](https://github.com/awslabs/aws-config-rules/tree/master/aws-config-conformance-packs).

~> **NOTE:** The account must have a Configuration Recorder with proper IAM permissions before the conformance pack will
successfully create or update. See also the
[`aws_config_configuration_recorder` resource](/docs/providers/aws/r/config_configuration_recorder.html).

## Example Usage

```hcl
resource "aws_config_conformance_pack" "test" {
  name          = "example"
  template_body = <<EOT
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

* `name` - (Required) The name of the conformance pack
* `template_s3_uri` - (Optional, required if `template_body` is not provided) Where to load the template from in S3 (ex: `s3://my-conformance-pack-bucket/packs/example-conformance-pack-template.yaml`).  This argument is not exported due to AWS API restrictions.
* `template_body` - (Optional, required if `template_s3_uri` is not provided) Body of the conformance pack template.  This argument is not exported due to AWS API restrictions.
* `input_parameters` - (Optional) Map of input parameters that is passed to the conformance pack template
* `delivery_s3_bucket` - (Optional) Amazon S3 bucket where AWS Config stores conformance pack templates
* `delivery_s3_key_prefix` - (Optional) Prefix for the Amazon S3 bucket where AWS Config stores conformance pack templates

## Attributes Reference

In addition to all arguments above (except `template_s3_uri` and `template_body`), the following attributes are exported:

* `arn` - Amazon Resource Name (ARN) of the conformance pack

## Import

Config Managed Rules can be imported using the name, e.g.

```
$ terraform import aws_config_conformance_pack.example example
```