// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package chatbot_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/service/chatbot"
	"github.com/aws/aws-sdk-go-v2/service/chatbot/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tfchatbot "github.com/hashicorp/terraform-provider-aws/internal/service/chatbot"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

const (
	testResourceSlackChannelConfiguration = "aws_chatbot_slack_channel_configuration.test"

	envSlackTeamID    = "CHATBOT_SLACK_TEAM_ID"
	envSlackChannelID = "CHATBOT_SLACK_CHANNEL_ID"
)

func TestAccChatbotSlackChannelConfiguration_serial(t *testing.T) {
	t.Parallel()

	testCases := map[string]func(t *testing.T){
		acctest.CtBasic:      testAccSlackChannelConfiguration_basic,
		acctest.CtDisappears: testAccSlackChannelConfiguration_disappears,
	}

	acctest.RunSerialTests1Level(t, testCases, 0)
}

func testAccSlackChannelConfiguration_basic(t *testing.T) {
	ctx := acctest.Context(t)

	var slackchannelconfiguration types.SlackChannelConfiguration
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	// The slack workspace must be created via the AWS Console. It cannot be created via APIs or Terraform.
	// Once it is created, export the name of the workspace in the env variable for this test
	teamID := acctest.SkipIfEnvVarNotSet(t, envSlackTeamID)
	channelID := acctest.SkipIfEnvVarNotSet(t, envSlackChannelID)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ChatbotServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSlackChannelConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSlackChannelConfigurationConfig_basic(rName, channelID, teamID),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSlackChannelConfigurationExists(ctx, testResourceSlackChannelConfiguration, &slackchannelconfiguration),
					resource.TestCheckResourceAttr(testResourceSlackChannelConfiguration, "configuration_name", rName),
					acctest.MatchResourceAttrGlobalARN(ctx, testResourceSlackChannelConfiguration, "chat_configuration_arn", "chatbot", regexache.MustCompile(fmt.Sprintf(`chat-configuration/slack-channel/%s`, rName))),
					resource.TestCheckResourceAttrPair(testResourceSlackChannelConfiguration, names.AttrIAMRoleARN, "aws_iam_role.test", names.AttrARN),
					resource.TestCheckResourceAttr(testResourceSlackChannelConfiguration, "slack_channel_id", channelID),
					resource.TestCheckResourceAttrSet(testResourceSlackChannelConfiguration, "slack_channel_name"),
					resource.TestCheckResourceAttr(testResourceSlackChannelConfiguration, "slack_team_id", teamID),
					resource.TestCheckResourceAttrSet(testResourceSlackChannelConfiguration, "slack_team_name"),
				),
			},
			{
				ResourceName:                         testResourceSlackChannelConfiguration,
				ImportState:                          true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(testResourceSlackChannelConfiguration, "chat_configuration_arn"),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "chat_configuration_arn",
			},
		},
	})
}

func testAccSlackChannelConfiguration_disappears(t *testing.T) {
	ctx := acctest.Context(t)

	var slackchannelconfiguration types.SlackChannelConfiguration
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	// The slack workspace must be created via the AWS Console. It cannot be created via APIs or Terraform.
	// Once it is created, export the name of the workspace in the env variable for this test
	teamID := acctest.SkipIfEnvVarNotSet(t, envSlackTeamID)
	channelID := acctest.SkipIfEnvVarNotSet(t, envSlackChannelID)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ChatbotServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSlackChannelConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSlackChannelConfigurationConfig_basic(rName, channelID, teamID),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSlackChannelConfigurationExists(ctx, testResourceSlackChannelConfiguration, &slackchannelconfiguration),
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tfchatbot.ResourceSlackChannelConfiguration, testResourceSlackChannelConfiguration),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckSlackChannelConfigurationDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).ChatbotClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_chatbot_slack_channel_configuration" {
				continue
			}

			_, err := tfchatbot.FindSlackChannelConfigurationByARN(ctx, conn, rs.Primary.Attributes["chat_configuration_arn"])

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return create.Error(names.Chatbot, create.ErrActionCheckingDestroyed, tfchatbot.ResNameSlackChannelConfiguration, rs.Primary.Attributes["chat_configuration_arn"], errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccCheckSlackChannelConfigurationExists(ctx context.Context, name string, slackchannelconfiguration *types.SlackChannelConfiguration) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.Chatbot, create.ErrActionCheckingExistence, tfchatbot.ResNameSlackChannelConfiguration, name, errors.New("not found"))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).ChatbotClient(ctx)

		output, err := tfchatbot.FindSlackChannelConfigurationByARN(ctx, conn, rs.Primary.Attributes["chat_configuration_arn"])

		if err != nil {
			return err
		}

		*slackchannelconfiguration = *output

		return nil
	}
}

func testAccPreCheck(ctx context.Context, t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).ChatbotClient(ctx)

	input := chatbot.DescribeSlackChannelConfigurationsInput{}
	_, err := conn.DescribeSlackChannelConfigurations(ctx, &input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}
	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccSlackChannelConfigurationConfig_basic(rName, channelID, teamID string) string {
	return fmt.Sprintf(`
data "aws_iam_policy_document" "assume_role" {
  statement {
    effect = "Allow"

    principals {
      type        = "Service"
      identifiers = ["chatbot.amazonaws.com"]
    }

    actions = ["sts:AssumeRole"]
  }
}

resource "aws_iam_role" "test" {
  name               = %[1]q
  assume_role_policy = data.aws_iam_policy_document.assume_role.json
}

resource "aws_chatbot_slack_channel_configuration" "test" {
  configuration_name = %[1]q
  iam_role_arn       = aws_iam_role.test.arn
  slack_channel_id   = %[2]q
  slack_team_id      = %[3]q

  tags = {
    Name = %[1]q
  }
}
`, rName, channelID, teamID)
}
