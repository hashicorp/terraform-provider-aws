---
layout: "aws"
page_title: "AWS: aws_config_remediation_configuration"
sidebar_current: "docs-aws-resource-config-remediation-configuration"
description: |-
  Provides an AWS Config Remediation Configuration.
---

# Resource: aws_config_remediation_configuration

Provides an AWS Config Remediation Configuration.

~> **Note:** Config Remediation Configuration requires an existing [Config Rule](/docs/providers/aws/r/config_config_rule.html) to be present. Use of `depends_on` is recommended (as shown below) to avoid race conditions.

## Example Usage

AWS managed rules can be used by setting the source owner to `AWS` and the source identifier to the name of the managed rule. More information about AWS managed rules can be found in the [AWS Config Developer Guide](https://docs.aws.amazon.com/config/latest/developerguide/evaluate-config_use-managed-rules.html).

```hcl
resource "aws_config_config_rule" "r" {
  name = "example"

  source {
    owner             = "AWS"
    source_identifier = "S3_BUCKET_VERSIONING_ENABLED"
  }

  depends_on = ["aws_config_configuration_recorder.foo"]
}

resource "aws_config_configuration_recorder" "foo" {
  name     = "example"
  role_arn = "${aws_iam_role.r.arn}"
}

resource "aws_iam_role" "r" {
  name = "my-awsconfig-role"

  assume_role_policy = <<POLICY
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "config.amazonaws.com"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
POLICY
}

resource "aws_iam_role_policy" "p" {
  name = "my-awsconfig-policy"
  role = "${aws_iam_role.r.id}"

  policy = <<POLICY
{
  "Version": "2012-10-17",
  "Statement": [
  	{
  		"Action": "config:Put*",
  		"Effect": "Allow",
  		"Resource": "*"

  	}
  ]
}
POLICY
}

resource "aws_sns_topic" "crc_topic" {
  name = "sns_topic_name"
}

resource "aws_config_remediation_configuration" "crc" {
  config_rule_name = "example"

  resource_type = ""
	target_id = "SSM_DOCUMENT"
	target_type = "AWS-PublishSNSNotification"
	target_version = "1"

	parameter {
		resource_value = "Message"
	}
	
	parameter {
		static_value {
			key   = "TopicArn"
			value = "${aws_sns_topic.crc_topic.arn}"
		}
	}
	
	parameter {
		static_value {
			key   = "AutomationAssumeRole"
			value = "${aws_iam_role.r.arn}"
		}
	}
}
```

## Argument Reference

The following arguments are supported:

* `config_rule_name` - (Required) The name of the AWS Config rule
* `resource_type` - (Optional) The type of a resource
* `target_id` - (Required) Target ID is the name of the public document
* `target_type` - (Required) The type of the target. Target executes remediation. For example, SSM document
* `target_version` - (Required) Version of the target. For example, version of the SSM document

### `parameters`

The value is either a dynamic (resource) value or a static value.
You must select either a dynamic value or a static value. 

* `resource_value` - (Optional) The value is dynamic and changes at run-time.
* `static_value` - (Optional) The value is static and does not change at run-time.

#### `static_value`

The value is static and does not change at run-time.

* `key` - (Required) The key of the parameter.
* `value` - (Required) The value of the parameter.

## Import

Remediation Configurations can be imported using the name config_rule_name, e.g.

```
$ terraform import aws_config_remediation_configuration.foo example
```
