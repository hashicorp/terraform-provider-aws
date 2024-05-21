// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ssmincidents_test

import (
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccResponsePlanDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rTitle := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_ssmincidents_response_plan.test"
	resourceName := "aws_ssmincidents_response_plan.test"

	snsTopic1 := "aws_sns_topic.topic1"
	snsTopic2 := "aws_sns_topic.topic2"

	displayName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	chatChannelTopic := "aws_sns_topic.channel_topic"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.SSMIncidentsEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SSMIncidentsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccResponsePlanDataSourceConfig_basic(
					rName,
					rTitle,
					snsTopic1,
					snsTopic2,
					displayName,
					chatChannelTopic,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, names.AttrName, dataSourceName, names.AttrName),
					resource.TestCheckResourceAttrPair(resourceName, "incident_template.0.title", dataSourceName, "incident_template.0.title"),
					resource.TestCheckResourceAttrPair(resourceName, "incident_template.0.impact", dataSourceName, "incident_template.0.impact"),
					resource.TestCheckResourceAttrPair(resourceName, "incident_template.0.dedupe_string", dataSourceName, "incident_template.0.dedupe_string"),
					resource.TestCheckResourceAttrPair(resourceName, "incident_template.0.summary", dataSourceName, "incident_template.0.summary"),
					resource.TestCheckResourceAttrPair(resourceName, "incident_template.0.incident_tags.%", dataSourceName, "incident_template.0.incident_tags.%"),
					resource.TestCheckResourceAttrPair(resourceName, "incident_template.0.incident_tags.a", dataSourceName, "incident_template.0.incident_tags.a"),
					resource.TestCheckResourceAttrPair(resourceName, "incident_template.0.incident_tags.b", dataSourceName, "incident_template.0.incident_tags.b"),
					resource.TestCheckTypeSetElemAttrPair(dataSourceName, "incident_template.0.notification_target.*.sns_topic_arn", snsTopic1, names.AttrARN),
					resource.TestCheckTypeSetElemAttrPair(dataSourceName, "incident_template.0.notification_target.*.sns_topic_arn", snsTopic2, names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrDisplayName, dataSourceName, names.AttrDisplayName),
					resource.TestCheckTypeSetElemAttrPair(dataSourceName, "chat_channel.0", chatChannelTopic, names.AttrARN),
					resource.TestCheckTypeSetElemAttrPair(
						resourceName,
						"engagements.0",
						dataSourceName,
						"engagements.0",
					),
					resource.TestCheckTypeSetElemAttrPair(
						dataSourceName,
						"action.0.ssm_automation.0.document_name",
						"aws_ssm_document.document",
						names.AttrName,
					),
					resource.TestCheckTypeSetElemAttrPair(
						dataSourceName,
						"action.0.ssm_automation.0.role_arn",
						"aws_iam_role.role",
						names.AttrARN,
					),
					resource.TestCheckResourceAttrPair(
						resourceName,
						"action.#",
						dataSourceName,
						"action.#",
					),
					resource.TestCheckResourceAttrPair(
						resourceName,
						"action.0.ssm_automation.#",
						dataSourceName,
						"action.0.ssm_automation.#",
					),
					resource.TestCheckResourceAttrPair(
						resourceName,
						"action.0.ssm_automation.0.document_version",
						dataSourceName,
						"action.0.ssm_automation.0.document_version",
					),
					resource.TestCheckResourceAttrPair(
						resourceName,
						"action.0.ssm_automation.0.target_account",
						dataSourceName,
						"action.0.ssm_automation.0.target_account",
					),
					resource.TestCheckResourceAttrPair(
						resourceName,
						"action.0.ssm_automation.0.parameter.#",
						dataSourceName,
						"action.0.ssm_automation.0.parameter.#",
					),
					resource.TestCheckResourceAttrPair(
						resourceName,
						"action.0.ssm_automation.0.parameter.0.name",
						dataSourceName,
						"action.0.ssm_automation.0.parameter.0.name",
					),
					resource.TestCheckResourceAttrPair(
						resourceName,
						"action.0.ssm_automation.0.parameter.0.values.#",
						dataSourceName,
						"action.0.ssm_automation.0.parameter.0.values.#",
					),
					resource.TestCheckResourceAttrPair(
						resourceName,
						"action.0.ssm_automation.0.parameter.0.values.0",
						dataSourceName,
						"action.0.ssm_automation.0.parameter.0.values.0",
					),
					resource.TestCheckResourceAttrPair(
						resourceName,
						"action.0.ssm_automation.0.parameter.0.values.1",
						dataSourceName,
						"action.0.ssm_automation.0.parameter.0.values.1",
					),
					resource.TestCheckResourceAttrPair(
						resourceName,
						"action.0.ssm_automation.0.dynamic_parameters.#",
						dataSourceName,
						"action.0.ssm_automation.0.dynamic_parameters.#",
					),
					resource.TestCheckResourceAttrPair(
						resourceName,
						"action.0.ssm_automation.0.dynamic_parameters.anotherKey",
						dataSourceName,
						"action.0.ssm_automation.0.dynamic_parameters.anotherKey",
					),
					resource.TestCheckResourceAttrPair(
						resourceName,
						"integration.#",
						dataSourceName,
						"integration.#",
					),
					resource.TestCheckResourceAttrPair(
						resourceName,
						"integration.0.pagerduty.#",
						dataSourceName,
						"integration.0.pagerduty.#",
					),
					resource.TestCheckResourceAttrPair(
						resourceName,
						"integration.0.pagerduty.0.name",
						dataSourceName,
						"integration.0.pagerduty.0.name",
					),
					resource.TestCheckResourceAttrPair(
						resourceName,
						"integration.0.pagerduty.0.service_id",
						dataSourceName,
						"integration.0.pagerduty.0.service_id",
					),
					resource.TestCheckResourceAttrPair(
						resourceName,
						"integration.0.pagerduty.0.secret_id",
						dataSourceName,
						"integration.0.pagerduty.0.secret_id",
					),
					resource.TestCheckResourceAttrPair(resourceName, acctest.CtTagsPercent, dataSourceName, acctest.CtTagsPercent),
					resource.TestCheckResourceAttrPair(resourceName, "tags.a", dataSourceName, "tags.a"),
					resource.TestCheckResourceAttrPair(resourceName, "tags.b", dataSourceName, "tags.b"),

					acctest.MatchResourceAttrGlobalARN(dataSourceName, names.AttrARN, "ssm-incidents", regexache.MustCompile(`response-plan/+.`)),
				),
			},
		},
	})
}

func testAccResponsePlanDataSourceConfig_basic(
	name,
	title,
	topic1,
	topic2,
	displayName,
	chatChannelTopic string) string {
	//lintignore:AWSAT003
	//lintignore:AWSAT005
	return fmt.Sprintf(`
resource "aws_sns_topic" "topic1" {}
resource "aws_sns_topic" "topic2" {}
resource "aws_sns_topic" "channel_topic" {}

resource "aws_ssmincidents_replication_set" "test_replication_set" {
  region {
    name = %[1]q
  }
}

resource "aws_iam_role" "role" {
  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Service": [
          "cloudformation.amazonaws.com"
        ]
      },
      "Action": [
        "sts:AssumeRole"
      ]
    }
  ]
}
EOF

  name = %[1]q
}

resource "aws_ssm_document" "document" {
  name          = %[1]q
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

resource "aws_ssmincidents_response_plan" "test" {
  name = %[2]q

  incident_template {
    title         = %[3]q
    impact        = "3"
    dedupe_string = "dedupe"
    summary       = "summary"

    incident_tags = {
      a = "tag1"
      b = ""
    }

    notification_target {
      sns_topic_arn = %[4]s
    }

    notification_target {
      sns_topic_arn = %[5]s
    }
  }

  display_name = %[6]q
  chat_channel = [%[7]s]

  engagements = ["arn:aws:ssm-contacts:us-east-2:111122223333:contact/test1"]

  action {
    ssm_automation {
      document_name    = aws_ssm_document.document.name
      role_arn         = aws_iam_role.role.arn
      document_version = "version1"
      target_account   = "RESPONSE_PLAN_OWNER_ACCOUNT"
      parameter {
        name   = "key"
        values = ["value1", "value2"]
      }
      dynamic_parameters = {
        anotherKey = "INVOLVED_RESOURCES"
      }
    }
  }

  #  Comment out integration section as the configured PagerDuty secretId is invalid and the test will fail,
  #  as we do not want to expose credentials to public repository.
  #  Tested locally and PagerDuty integration work with response plan.
  #  integration {
  #    pagerduty {
  #      name = "pagerduty-test-terraform"
  #      service_id = "PNDIQ3N"
  #      secret_id = "PagerdutyPoshchuSecret"
  #    }
  #  }

  tags = {
    a = "tag1"
    b = ""
  }

  depends_on = [aws_ssmincidents_replication_set.test_replication_set]
}

data "aws_ssmincidents_response_plan" "test" {
  arn = aws_ssmincidents_response_plan.test.arn
}
`,
		acctest.Region(),
		name,
		title,
		topic1+".arn",
		topic2+".arn",
		displayName,
		chatChannelTopic+".arn",
	)
}
