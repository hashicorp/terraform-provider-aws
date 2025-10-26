---
subcategory: "SSM (Systems Manager)"
layout: "aws"
page_title: "AWS: aws_ssm_send_command"
description: |-
  Executes commands on EC2 instances using AWS Systems Manager Run Command.
---

# Action: aws_ssm_send_command

~> **Note:** `aws_ssm_send_command` is in alpha. Its interface and behavior may change as the feature evolves, and breaking changes are possible. It is offered as a technical preview without compatibility guarantees until Terraform 1.14 is generally available.

!> **Warning:** This action may cause unintended consequences. When triggered, the `aws_ssm_send_command` action executes commands on EC2 instances, which can modify system state, install software, or perform other operations. Use caution—this preview action should be limited to development environments or carefully controlled production scenarios.

Executes commands on EC2 instances using AWS Systems Manager Run Command. This action sends a command to one or more instances and waits for execution to complete.

For information about AWS Systems Manager Run Command, see the [AWS Systems Manager User Guide](https://docs.aws.amazon.com/systems-manager/latest/userguide/execute-remote-commands.html). For specific information about the SendCommand API, see the [SendCommand](https://docs.aws.amazon.com/systems-manager/latest/APIReference/API_SendCommand.html) page in the AWS Systems Manager API Reference.

~> **Note:** This action requires that target instances are managed by AWS Systems Manager and have the SSM Agent installed and running.

## Example Usage

### Basic Usage

```terraform
resource "aws_instance" "example" {
  ami           = data.aws_ami.amazon_linux.id
  instance_type = "t3.micro"
  iam_instance_profile = aws_iam_instance_profile.ssm.name

  tags = {
    Name = "example-instance"
  }
}

action "aws_ssm_send_command" "example" {
  config {
    instance_ids  = [aws_instance.example.id]
    document_name = "AWS-RunShellScript"
    parameters = {
      commands = ["echo 'Hello World'", "uptime"]
    }
  }
}
```

### With Parameters

```terraform
action "aws_ssm_send_command" "update_packages" {
  config {
    instance_ids  = [aws_instance.web.id]
    document_name = "AWS-RunShellScript"
    parameters = {
      commands = [
        "sudo yum update -y",
        "sudo systemctl restart httpd"
      ]
    }
    timeout = 600
  }
}
```

### Multiple Instances with S3 Output

```terraform
action "aws_ssm_send_command" "deploy" {
  config {
    instance_ids = [
      aws_instance.web1.id,
      aws_instance.web2.id,
      aws_instance.web3.id
    ]
    document_name      = "AWS-RunShellScript"
    output_s3_bucket   = aws_s3_bucket.logs.id
    parameters = {
      commands = [
        "cd /var/www/html",
        "git pull origin main",
        "sudo systemctl reload nginx"
      ]
    }
    timeout = 900
  }
}
```

### Deployment Trigger

```terraform
resource "aws_instance" "app_server" {
  ami           = data.aws_ami.amazon_linux.id
  instance_type = "t3.micro"
  iam_instance_profile = aws_iam_instance_profile.ssm.name

  tags = {
    Name = "app-server"
  }
}

action "aws_ssm_send_command" "deploy_app" {
  config {
    instance_ids  = [aws_instance.app_server.id]
    document_name = "AWS-RunShellScript"
    parameters = {
      commands = [
        "aws s3 cp s3://my-app-bucket/app.zip /tmp/",
        "unzip -o /tmp/app.zip -d /opt/app",
        "sudo systemctl restart app"
      ]
    }
    timeout = 1200
  }
}

resource "terraform_data" "deployment_trigger" {
  input = var.app_version

  lifecycle {
    action_trigger {
      events  = [after_create, after_update]
      actions = [action.aws_ssm_send_command.deploy_app]
    }
  }

  depends_on = [aws_instance.app_server]
}
```

## Argument Reference

This action supports the following arguments:

* `instance_ids` - (Required) List of EC2 instance IDs to execute the command on. All instances must be managed by AWS Systems Manager.
* `document_name` - (Required) Name of the SSM document to execute. Can be an AWS-provided document (e.g., `AWS-RunShellScript`, `AWS-RunPowerShellScript`) or a custom document.
* `parameters` - (Optional) Parameters to pass to the command document. The structure depends on the document being executed. For `AWS-RunShellScript`, use `commands` as a list of shell commands.
* `output_s3_bucket` - (Optional) S3 bucket name to store command output. The SSM service must have permission to write to this bucket.
* `region` - (Optional) Region where this action should be [run](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `timeout` - (Optional) Timeout in seconds to wait for command execution to complete. Must be between 60 and 7200 seconds. Default: `1800`.

## Behavior

The action performs the following steps:

1. Validates that all specified instances exist and are managed by SSM
2. Validates that the specified document exists
3. Sends the command to all instances
4. Polls command status every 5 seconds until all invocations complete
5. Checks exit codes and reports any failures

If any instance fails to execute the command successfully, the action fails and reports the error details including exit codes and stderr output.

## Prerequisites

To use this action, ensure:

* Target EC2 instances have the SSM Agent installed and running
* Instances have an IAM instance profile with the `AmazonSSMManagedInstanceCore` policy
* Instances are registered with Systems Manager (visible in the SSM console)
* If using `output_s3_bucket`, the SSM service has permission to write to the bucket

## Common Documents

AWS provides several built-in documents for common tasks:

* `AWS-RunShellScript` - Run shell commands on Linux instances
* `AWS-RunPowerShellScript` - Run PowerShell commands on Windows instances
* `AWS-ConfigureAWSPackage` - Install or update AWS packages
* `AWS-UpdateSSMAgent` - Update the SSM Agent
* `AWS-RunPatchBaseline` - Apply patch baselines

For a complete list, see the [AWS Systems Manager documents reference](https://docs.aws.amazon.com/systems-manager/latest/userguide/sysman-ssm-docs.html).
