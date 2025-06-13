// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package chatbot_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

const (
	// Environment variables for data source testing
	envSlackConfigurationName = "CHATBOT_SLACK_CONFIGURATION_NAME"
	envSlackTeamName          = "CHATBOT_SLACK_TEAM_NAME"
)

func TestAccChatbotSlackChannelConfigurationDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_chatbot_slack_channel_configuration.test"

	// The slack workspace and configuration must be created via the AWS Console.
	// They cannot be created via APIs or Terraform.
	// Export the configuration details in env variables for this test
	teamID := acctest.SkipIfEnvVarNotSet(t, envSlackTeamID)
	channelID := acctest.SkipIfEnvVarNotSet(t, envSlackChannelID)
	configurationName := acctest.SkipIfEnvVarNotSet(t, envSlackConfigurationName)
	teamName := acctest.SkipIfEnvVarNotSet(t, envSlackTeamName)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ChatbotServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccSlackChannelConfigurationDataSourceConfig_basic(configurationName),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Static assertions - values from environment variables
					resource.TestCheckResourceAttr(dataSourceName, "configuration_name", configurationName),
					resource.TestCheckResourceAttr(dataSourceName, "slack_channel_id", channelID),
					resource.TestCheckResourceAttr(dataSourceName, "slack_team_id", teamID),
					resource.TestCheckResourceAttr(dataSourceName, "slack_team_name", teamName),

					// Dynamic assertions - values set by AWS, just verify they exist
					resource.TestCheckResourceAttrSet(dataSourceName, "chat_configuration_arn"),
					resource.TestCheckResourceAttrSet(dataSourceName, names.AttrIAMRoleARN),
					resource.TestCheckResourceAttrSet(dataSourceName, "logging_level"),
					resource.TestCheckResourceAttrSet(dataSourceName, "slack_channel_name"),
					resource.TestCheckResourceAttrSet(dataSourceName, "user_authorization_required"),
					resource.TestCheckResourceAttrSet(dataSourceName, names.AttrState),
					resource.TestCheckResourceAttrSet(dataSourceName, "sns_topic_arns.#"),
					resource.TestCheckResourceAttrSet(dataSourceName, "tags.%"),

					// ARN pattern validation - ensures correct format with expected configuration name
					acctest.MatchResourceAttrGlobalARN(ctx, dataSourceName, "chat_configuration_arn", "chatbot", regexache.MustCompile(`chat-configuration/slack-channel/`+regexp.QuoteMeta(configurationName)+`$`)),
				),
			},
		},
	})
}

func TestAccChatbotSlackChannelConfigurationDataSource_tags(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_chatbot_slack_channel_configuration.test"

	configurationName := acctest.SkipIfEnvVarNotSet(t, envSlackConfigurationName)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ChatbotServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccSlackChannelConfigurationDataSourceConfig_tags(configurationName),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Core resource identification
					resource.TestCheckResourceAttrSet(dataSourceName, "chat_configuration_arn"),
					acctest.MatchResourceAttrGlobalARN(ctx, dataSourceName, "chat_configuration_arn", "chatbot", regexache.MustCompile(`chat-configuration/slack-channel/`+regexp.QuoteMeta(configurationName)+`$`)),

					// Tag validation - verify tags are populated by transparent tagging
					resource.TestCheckResourceAttrSet(dataSourceName, "tags.%"),
					// Note: Specific tag values depend on the existing AWS resource configuration
					// This test validates that the transparent tagging system correctly retrieves tags
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

func testAccSlackChannelConfigurationDataSourceConfig_basic(configurationName string) string {
	return fmt.Sprintf(`
data "aws_caller_identity" "current" {}

data "aws_chatbot_slack_channel_configuration" "test" {
  chat_configuration_arn = "arn:aws:chatbot::${data.aws_caller_identity.current.account_id}:chat-configuration/slack-channel/%[1]s"
}
`, configurationName)
}

func testAccSlackChannelConfigurationDataSourceConfig_tags(configurationName string) string {
	return fmt.Sprintf(`
data "aws_caller_identity" "current" {}

data "aws_chatbot_slack_channel_configuration" "test" {
  chat_configuration_arn = "arn:aws:chatbot::${data.aws_caller_identity.current.account_id}:chat-configuration/slack-channel/%[1]s"
}
`, configurationName)
}

func testAccSlackChannelConfigurationDataSourceConfig_notFound() string {
	return `
data "aws_caller_identity" "current" {}

data "aws_chatbot_slack_channel_configuration" "test" {
  chat_configuration_arn = "arn:aws:chatbot::${data.aws_caller_identity.current.account_id}:chat-configuration/slack-channel/does-not-exist"
}
`
}
