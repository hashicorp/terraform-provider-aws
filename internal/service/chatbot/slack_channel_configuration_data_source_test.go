// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package chatbot_test

import (
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccChatbotSlackChannelConfigurationDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_chatbot_slack_channel_configuration.test"
	resourceName := "aws_chatbot_slack_channel_configuration.test"

	// The slack workspace must be created via the AWS Console. It cannot be created via APIs or Terraform.
	// Once it is created, export the workspace details in the env variables for this test
	teamID := acctest.SkipIfEnvVarNotSet(t, envSlackTeamID)
	channelID := acctest.SkipIfEnvVarNotSet(t, envSlackChannelID)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ChatbotServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSlackChannelConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSlackChannelConfigurationDataSourceConfig_basic(rName, channelID, teamID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "chat_configuration_arn", resourceName, "chat_configuration_arn"),
					resource.TestCheckResourceAttrPair(dataSourceName, "configuration_name", resourceName, "configuration_name"),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrIAMRoleARN, resourceName, names.AttrIAMRoleARN),
					resource.TestCheckResourceAttrPair(dataSourceName, "logging_level", resourceName, "logging_level"),
					resource.TestCheckResourceAttrPair(dataSourceName, "slack_channel_id", resourceName, "slack_channel_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "slack_channel_name", resourceName, "slack_channel_name"),
					resource.TestCheckResourceAttrPair(dataSourceName, "slack_team_id", resourceName, "slack_team_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "slack_team_name", resourceName, "slack_team_name"),
					resource.TestCheckResourceAttrPair(dataSourceName, "user_authorization_required", resourceName, "user_authorization_required"),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrState, resourceName, names.AttrState),
					resource.TestCheckResourceAttrPair(dataSourceName, "sns_topic_arns", resourceName, "sns_topic_arns"),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrTags, resourceName, names.AttrTags),
					// Use GlobalARN matcher as Chatbot ARNs are global resources
					acctest.MatchResourceAttrGlobalARN(ctx, dataSourceName, "chat_configuration_arn", "chatbot", regexache.MustCompile(fmt.Sprintf(`chat-configuration/slack-channel/%s$`, rName))),
				),
			},
		},
	})
}

func TestAccChatbotSlackChannelConfigurationDataSource_tags(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_chatbot_slack_channel_configuration.test"
	resourceName := "aws_chatbot_slack_channel_configuration.test"

	teamID := acctest.SkipIfEnvVarNotSet(t, envSlackTeamID)
	channelID := acctest.SkipIfEnvVarNotSet(t, envSlackChannelID)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ChatbotServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSlackChannelConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSlackChannelConfigurationDataSourceConfig_tags(rName, channelID, teamID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "chat_configuration_arn", resourceName, "chat_configuration_arn"),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrTags, resourceName, names.AttrTags),
					resource.TestCheckResourceAttr(dataSourceName, acctest.CtTagsPercent, "2"),
					resource.TestCheckResourceAttr(dataSourceName, acctest.CtTagsKey1, acctest.CtValue1),
					resource.TestCheckResourceAttr(dataSourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func TestAccChatbotSlackChannelConfigurationDataSource_notFound(t *testing.T) {
	ctx := acctest.Context(t)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ChatbotServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccSlackChannelConfigurationDataSourceConfig_notFound(),
				ExpectError: regexache.MustCompile(`empty result`),
			},
		},
	})
}

func testAccSlackChannelConfigurationDataSourceConfig_basic(rName, channelID, teamID string) string {
	return acctest.ConfigCompose(testAccSlackChannelConfigurationConfig_basic(rName, channelID, teamID), `
data "aws_chatbot_slack_channel_configuration" "test" {
  chat_configuration_arn = aws_chatbot_slack_channel_configuration.test.chat_configuration_arn
}
`)
}

func testAccSlackChannelConfigurationDataSourceConfig_tags(rName, channelID, teamID string) string {
	return acctest.ConfigCompose(testAccSlackChannelConfigurationConfig_tags1(rName, channelID, teamID, acctest.CtKey1, acctest.CtValue1), `
data "aws_chatbot_slack_channel_configuration" "test" {
  chat_configuration_arn = aws_chatbot_slack_channel_configuration.test.chat_configuration_arn
}
`)
}

func testAccSlackChannelConfigurationDataSourceConfig_notFound() string {
	return `
data "aws_chatbot_slack_channel_configuration" "test" {
  chat_configuration_arn = "arn:aws:chatbot::123456789012:chat-configuration/slack-channel/does-not-exist"
}
`
}
