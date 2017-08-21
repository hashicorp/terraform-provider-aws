---
layout: "aws"
page_title: "AWS: aws_ssm_association"
sidebar_current: "docs-aws-resource-ssm-association"
description: |-
  Associates an SSM Document to an instance.
---

# aws_ssm_association

Associates an SSM Document to an instance.

## Example Usage

```hcl
resource "aws_security_group" "tf_test_foo" {
  name        = "tf_test_foo"
  description = "foo"

  ingress {
    protocol    = "icmp"
    from_port   = -1
    to_port     = -1
    cidr_blocks = ["0.0.0.0/0"]
  }
}

resource "aws_instance" "foo" {
  # eu-west-1
  ami               = "ami-f77ac884"
  availability_zone = "eu-west-1a"
  instance_type     = "t2.small"
  security_groups   = ["${aws_security_group.tf_test_foo.name}"]
}

resource "aws_ssm_document" "foo_document" {
  name = "test_document_association-%s"

  content = <<DOC
  {
    "schemaVersion": "1.2",
    "description": "Check ip configuration of a Linux instance.",
    "parameters": {

    },
    "runtimeConfig": {
      "aws:runShellScript": {
        "properties": [
          {
            "id": "0.aws:runShellScript",
            "runCommand": ["ifconfig"]
          }
        ]
      }
    }
  }
DOC
}

resource "aws_ssm_association" "foo" {
  name        = "test_document_association-%s"
  instance_id = "${aws_instance.foo.id}"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the SSM document to apply.
* `instance_id` - (Optional) The instance id to apply an SSM document to.
* `parameters` - (Optional) Additional parameters to pass to the SSM document.
* `targets` - (Optional) The targets (either instances or tags). Instances are specified using Key=instanceids,Values=instanceid1,instanceid2. Tags are specified using Key=tag name,Values=tag value. Only 1 target is currently supported by AWS.
* `schedule_expression` - (Optional) A cron expression when the association will be applied to the target(s).
* `output_location` - (Optional) An output location block. OutputLocation documented below.
* `document_version` - (Optional) The document version you want to associate with the target(s). Can be a specific version or the default version.

Output Location (`output_location`) is an S3 bucket where you want to store the results of this association:

* `s3_bucket_name` - (Required) The S3 bucket name.
* `s3_key_prefix` - (Optional) The S3 bucket prefix. Results stored in the root if not configured.

## Attributes Reference

The following attributes are exported:

* `name` - The name of the SSM document to apply.
* `instance_ids` - The instance id that the SSM document was applied to.
* `parameters` - Additional parameters passed to the SSM document.
