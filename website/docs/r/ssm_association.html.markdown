---
layout: "aws"
page_title: "AWS: aws_ssm_association"
sidebar_current: "docs-aws-resource-ssm-association"
description: |-
  Associates an SSM Document to an instance or EC2 tag.
---

# aws_ssm_association

Associates an SSM Document to an instance or EC2 tag.

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
  name          = "test_document_association-%s"
  document_type = "Command"

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
* `association_name` - (Optional) The descriptive name for the association.
* `document_version` - (Optional) The document version you want to associate with the target(s). Can be a specific version or the default version.
* `instance_id` - (Optional) The instance ID to apply an SSM document to.
* `output_location` - (Optional) An output location block. Output Location is documented below.
* `parameters` - (Optional) A block of arbitrary string parameters to pass to the SSM document.
* `schedule_expression` - (Optional) A cron expression when the association will be applied to the target(s).
* `targets` - (Optional) A block containing the targets of the SSM association. Targets are documented below. AWS currently supports a maximum of 5 targets.

Output Location (`output_location`) is an S3 bucket where you want to store the results of this association:

* `s3_bucket_name` - (Required) The S3 bucket name.
* `s3_key_prefix` - (Optional) The S3 bucket prefix. Results stored in the root if not configured.

Targets specify what instance IDs or tags to apply the document to and has these keys:

* `key` - (Required) Either `InstanceIds` or `tag:Tag Name` to specify an EC2 tag.
* `values` - (Required) A list of instance IDs or tag values. AWS currently limits this to 1 target value.

## Attributes Reference

The following attributes are exported:

* `name` - The name of the SSM document to apply.
* `instance_ids` - The instance id that the SSM document was applied to.
* `parameters` - Additional parameters passed to the SSM document.
