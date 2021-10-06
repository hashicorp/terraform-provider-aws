---
name: 🐛 Bug Report
about: unexpected behavior of Running SSM Document 🤔.

---

<!---
Please note the following potential times when an issue might be in Terraform core:

* [Configuration Language](https://www.terraform.io/docs/configuration/index.html) or resource ordering issues
* [State](https://www.terraform.io/docs/state/index.html) and [State Backend](https://www.terraform.io/docs/backends/index.html) issues
* [Provisioner](https://www.terraform.io/docs/provisioners/index.html) issues
* [Registry](https://registry.terraform.io/) issues
* Spans resources across multiple providers

If you are running into one of these scenarios, we recommend opening an issue in the [Terraform core repository](https://github.com/hashicorp/terraform/) instead.
--->

<!--- Please keep this note for the community --->

### Community Note

* Please vote on this issue by adding a 👍 [reaction](https://blog.github.com/2016-03-10-add-reactions-to-pull-requests-issues-and-comments/) to the original issue to help the community and maintainers prioritize this request
* Please do not leave "+1" or other comments that do not add relevant new information or questions, they generate extra noise for issue followers and do not help prioritize the request
* If you are interested in working on this issue or have submitted a pull request, please leave a comment

<!--- Thank you for keeping this note for the community --->

### Terraform CLI and Terraform AWS Provider Version
Terraform verion: v0.14.11
provider registry.terraform.io/hashicorp/aws v3.61.0

<!--- Please run `terraform -v` to show the Terraform core version and provider version(s). If you are not running the latest version of Terraform or the provider, please upgrade because your issue may have already been fixed. [Terraform documentation on provider versioning](https://www.terraform.io/docs/configuration/providers.html#provider-versions). --->

### Affected Resource(s)

<!--- Please list the affected resources and data sources. --->

* aws_ssm_document
* aws_ssm_association

### Terraform Configuration Files

<!--- Information about code formatting: https://help.github.com/articles/basic-writing-and-formatting-syntax/#quoting-code --->

Please include all Terraform configurations required to reproduce the bug. Bug reports without a functional reproduction may be closed without investigation.

resource "aws_ssm_document" "SSM-test-DOC1" {
  name          = "SSM-DOC-test1"
  document_type = "Command"
  target_type = "/AWS::EC2::Instance"
  content       = <<DOC
  {
    "schemaVersion": "2.2",
    "description": "Execute scripts stored in a remote location. The following remote locations are currently supported: GitHub (public and private) and Amazon S3 (S3). The following script types are currently supported: #! support on Linux and file associations on Windows.",
    "parameters": {
      "sourceType": {
        "description": "(Required) Specify the source type.",
        "type": "String",
        "default": "S3"
      },
      "sourceinf": {
          "description": "(Required) The information required to retrieve the content from the required source.",
          "type": "StringMap",
          "displayType": "textarea",
          "default":  {
            "path": "<s3 bucket path>"
          }
      },
      "commandLine": {
        "description": "(Required) Specify the command line to be executed. The following formats of commands can be run: 'pythonMainFile.py argument1 argument2', 'ansible-playbook -i \"localhost,\" -c local example.yml'",
        "type": "String",
        "default": "./getting.sh" # this is the file I want ot download from S3 and run on the instance specified at 'aws_ssm_association'
      },

      "workingDirectory": {
        "type": "String",
        "default": "home/ec2-user",
        "description": "(Optional) The path where the content will be downloaded and executed from on your instance.",
        "maxChars": 4096
      }
    
    },
    "mainSteps": [
      {
        "action": "aws:downloadContent",
        "name": "downloadContent",
        "inputs": {
          "sourceType": "{{ sourceType }}",
          "sourceInfo": "{{ sourceinf }}",
          "destinationPath": "{{ workingDirectory }}"
        }
      },
      {
        "action": "aws:runShellScript",
        "name": "runShellScript",
        "inputs": {
          "runCommand": [
            "directory=$(pwd)",
            "export PATH=$PATH:$directory",
            "{{commandLine}}",
            "touch fff.txt",
            "cp getting.sh getting3.sh",
            "echo howqahe"
          ],
          "workingDirectory": "{{ workingDirectory }}"

        }
      }
    ]
  }
DOC
}

resource "aws_ssm_association" "ec2ssm1" {
  name = aws_ssm_document.SSM-test-DOC1.name

  targets {
    key    = "InstanceIds"
    values = [
      "<ec2-id>",
      ]
  }
}



### Debug Output

after running terraform apply I got:
aws_ssm_document.SSM-test-DOC1: Refreshing state... [id=SSM-DOC-test1]
aws_ssm_association.ec2ssm1: Refreshing state... [id=6c19c445-e829-41b1-afa1-0fea2c219582]
aws_ssm_document.SSM-test-DOC1: Modifying... [id=SSM-DOC-test1]
aws_ssm_document.SSM-test-DOC1: Modifications complete after 5s [id=SSM-DOC-test1]
                        

Apply complete! Resources: 1 added, 0 changed, 0 destroyed.
-----------
BUT:
into the SSM -> Run Command -> command History
  the command failed due to failure in downloading the s3 bucket file specified in the 'sourceinf' block although it exists there!
                        
### Panic Output

NONE

### Expected Behavior

to a status of Success in the SSM -> Run Command -> command History
(which I get when running the command manually in the console using the Run Commnad option)
  
### Actual Behavior

downloading the S3 script file failed -> running the command failed, and got the following error:

invalid format in plugin properties map[destinationPath:home/ec2-user sourceInfo:map[path:https://test-ssm-document-2.s3.amazonaws.com/getting.sh] sourceType:S3];

error json: cannot unmarshal object into Go struct field DownloadContentPlugin.sourceInfo of type string

  
### Steps to Reproduce


1. `create a script file and upload it to S3 bucket`
2. `create an EC2 instance in the same region of the S3 bucket with SSM role to run SSM command on it`
3. `update the ec2 id and s3 file path in the .tf file`
4. `terraform apply`
5. `in the terminal everything should work fine`
6. `log into the aws console -> SSM -> Run Command -> Command History`
7. `Status of the last Command will be "failed"`

### Important Factoids

None

### References

None

* #0000
