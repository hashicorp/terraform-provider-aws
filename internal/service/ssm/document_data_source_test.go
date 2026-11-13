// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package ssm_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccSSMDocumentDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_ssm_document.test"
	resourceName := "aws_ssm_document.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SSMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDocumentDataSourceConfig_basic(rName, "JSON"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrARN, resourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrName, resourceName, names.AttrName),
					resource.TestCheckResourceAttrPair(dataSourceName, "document_format", resourceName, "document_format"),
					resource.TestCheckResourceAttr(dataSourceName, "document_version", "1"),
					resource.TestCheckResourceAttrPair(dataSourceName, "document_type", resourceName, "document_type"),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrContent, resourceName, names.AttrContent),
				),
			},
			{
				Config: testAccDocumentDataSourceConfig_basic(rName, "YAML"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrARN, resourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrName, resourceName, names.AttrName),
					resource.TestCheckResourceAttr(dataSourceName, "document_format", "YAML"),
					resource.TestCheckResourceAttr(dataSourceName, "document_version", "1"),
					resource.TestCheckResourceAttrPair(dataSourceName, "document_type", resourceName, "document_type"),
					resource.TestCheckResourceAttrSet(dataSourceName, names.AttrContent),
				),
			},
		},
	})
}

func TestAccSSMDocumentDataSource_basicAutomation(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_ssm_document.test"
	resourceName := "aws_ssm_document.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SSMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDocumentDataSourceConfig_basicAutomation(rName, "JSON"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrARN, resourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrName, resourceName, names.AttrName),
					resource.TestCheckResourceAttrPair(dataSourceName, "document_format", resourceName, "document_format"),
					resource.TestCheckResourceAttr(dataSourceName, "document_version", "1"),
					resource.TestCheckResourceAttrPair(dataSourceName, "document_type", resourceName, "document_type"),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrContent, resourceName, names.AttrContent),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrName, resourceName, names.AttrName),
				),
			},
			{
				Config: testAccDocumentDataSourceConfig_basicAutomation(rName, "YAML"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrARN, resourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrName, resourceName, names.AttrName),
					resource.TestCheckResourceAttr(dataSourceName, "document_format", "YAML"),
					resource.TestCheckResourceAttr(dataSourceName, "document_version", "1"),
					resource.TestCheckResourceAttrPair(dataSourceName, "document_type", resourceName, "document_type"),
					resource.TestCheckResourceAttrSet(dataSourceName, names.AttrContent),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrName, resourceName, names.AttrName),
				),
			},
		},
	})
}

func TestAccSSMDocumentDataSource_managed(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_ssm_document.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SSMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDocumentDataSourceConfig_managed(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, names.AttrName, "AWS-StartEC2Instance"),
					resource.TestCheckResourceAttr(dataSourceName, names.AttrARN, "AWS-StartEC2Instance"),
				),
			},
		},
	})
}

func testAccDocumentDataSourceConfig_basic(rName, documentFormat string) string {
	return fmt.Sprintf(`
resource "aws_ssm_document" "test" {
  name          = %[1]q
  document_type = "Command"

  content = <<DOC
{
  "schemaVersion": "1.2",
  "description": "Check ip configuration of a Linux instance.",
  "parameters": {},
  "runtimeConfig": {
    "aws:runShellScript": {
      "properties": [
        {
          "id": "0.aws:runShellScript",
          "runCommand": [
            "ifconfig"
          ]
        }
      ]
    }
  }
}
DOC
}

# data.aws_ssm_documen.test.document_format can change regardless of the type of aws_ssm_document.test.document_format
# because api can return different representation of content, for example: real content is json, but can return YAML on demand.
# that's why we can give both JSON and YAML on same aws_ssm_document
data "aws_ssm_document" "test" {
  name            = aws_ssm_document.test.name
  document_format = %[2]q
}
`, rName, documentFormat)
}

func testAccDocumentDataSourceConfig_basicAutomation(rName, documentFormat string) string {
	return acctest.ConfigCompose(acctest.ConfigLatestAmazonLinux2HVMEBSX8664AMI(),
		fmt.Sprintf(`
resource "aws_iam_instance_profile" "ssm_profile" {
  name = %[1]q
  role = aws_iam_role.ssm_role.name
}

data "aws_partition" "current" {}

resource "aws_iam_role" "ssm_role" {
  name = %[1]q
  path = "/"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "ec2.${data.aws_partition.current.dns_suffix}"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF
}

resource "aws_ssm_document" "test" {
  name          = %[1]q
  document_type = "Automation"

  content = <<DOC
{
  "description": "Systems Manager Automation Demo",
  "schemaVersion": "0.3",
  "assumeRole": "${aws_iam_role.ssm_role.arn}",
  "mainSteps": [
    {
      "name": "startInstances",
      "action": "aws:runInstances",
      "timeoutSeconds": 1200,
      "maxAttempts": 1,
      "onFailure": "Abort",
      "inputs": {
        "ImageId": "${data.aws_ami.amzn2-ami-minimal-hvm-ebs-x86_64.id}",
        "InstanceType": "t2.small",
        "MinInstanceCount": 1,
        "MaxInstanceCount": 1,
        "IamInstanceProfileName": "${aws_iam_instance_profile.ssm_profile.name}"
      }
    },
    {
      "name": "stopInstance",
      "action": "aws:changeInstanceState",
      "maxAttempts": 1,
      "onFailure": "Continue",
      "inputs": {
        "InstanceIds": [
          "{{ startInstances.InstanceIds }}"
        ],
        "DesiredState": "stopped"
      }
    },
    {
      "name": "terminateInstance",
      "action": "aws:changeInstanceState",
      "maxAttempts": 1,
      "onFailure": "Continue",
      "inputs": {
        "InstanceIds": [
          "{{ startInstances.InstanceIds }}"
        ],
        "DesiredState": "terminated"
      }
    }
  ]
}
DOC
}

# data.aws_ssm_documen.test.document_format can change regardless of the type of aws_ssm_document.test.document_format
# because api can return different representation of content, for example: real content is json, but can return YAML on demand.
# that's why we can give both JSON and YAML on same aws_ssm_document
data "aws_ssm_document" "test" {
  name            = aws_ssm_document.test.name
  document_format = %[2]q
}
`, rName, documentFormat))
}

func testAccDocumentDataSourceConfig_managed() string {
	return `
data "aws_ssm_document" "test" {
  name = "AWS-StartEC2Instance"
}
`
}
