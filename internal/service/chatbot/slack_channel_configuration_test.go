// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package chatbot_test

import (
	"context"
	"errors"
	"fmt"
	"os"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	awstypes "github.com/aws/aws-sdk-go-v2/service/chatbot/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"

	tfchatbot "github.com/hashicorp/terraform-provider-aws/internal/service/chatbot"
)

func getEnvironmentVariables(t *testing.T) (string, string) {
	// The slack workspace must be created via the AWS Console. It cannot be created via APIs or Terraform.
	// Once it is created, export the name of the workspace in the env variable for this test
	workspaceNameKey := "CHATBOT_SLACK_WORKSPACE_NAME"
	workspaceName := os.Getenv(workspaceNameKey)

	if workspaceName == "" {
		t.Skipf("Environment variable %s is not set", workspaceNameKey)
	}

	// This is the ID of the slack channel. The ID of the Slack channel.
	// To get the ID, open Slack, right click on the channel name in the left pane, then choose Copy Link.
	// The channel ID is the 9-character string at the end of the URL. For example, ABCBBLZZZ.
	slackChannelIdKey := "CHATBOT_SLACK_CHANNEL_ID"
	slackChannelId := os.Getenv(slackChannelIdKey)
	if slackChannelId == "" {
		t.Skipf("Environment variable %s is not set", slackChannelIdKey)
	}

	return workspaceName, slackChannelId
}

func TestAccChatbotSlackChannelConfiguration_basic(t *testing.T) {
	ctx := acctest.Context(t)
	// TIP: This is a long-running test guard for tests that run longer than
	// 300s (5 min) generally.
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	// Get the necessary environment variables to run this test.
	workspaceName, slackChannelId := getEnvironmentVariables(t)

	var slackchannelconfiguration awstypes.SlackChannelConfiguration
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_chatbot_slack_channel_configuration.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ChatbotServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSlackChannelConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSlackChannelConfigurationConfig_basic(rName, workspaceName, slackChannelId),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSlackChannelConfigurationExists(ctx, resourceName, &slackchannelconfiguration),
					resource.TestCheckResourceAttr(resourceName, "configuration_name", rName),
					resource.TestCheckResourceAttr(resourceName, "guardrail_policy_arns.#", acctest.Ct1),
					acctest.MatchResourceAttrGlobalARN(resourceName, "iam_role_arn", "iam", regexache.MustCompile("role/.+-1$")),
					resource.TestCheckResourceAttr(resourceName, "logging_level", "NONE"),
					resource.TestCheckResourceAttr(resourceName, "slack_channel_id", slackChannelId),
					//SlackChannelName is not being tested as the Slack Channel Name (if not provided as input) is not always available.
					//The service queries Slack to get the channel name, but if the Slack Channel is a private channle without @aws both installed in the channel, then the SlackChannelName will not be available for the service.
					//The only tests where SlackChannelName is validated is when the SlackChannelName is provided as an input attribute
					//resource.TestCheckResourceAttrSet(resourceName, "slack_channel_name"),
					resource.TestCheckResourceAttrSet(resourceName, "slack_team_id"),
					resource.TestCheckResourceAttr(resourceName, "sns_topic_arns.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "user_authorization_required", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
				),
				ExpectNonEmptyPlan: false,
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"slack_channel_name", // This attribute is not always returned by the Create API immediately. And sometimes never. See the comment in TestAccChatbotSlackChannelConfiguration_basic
				},
			},
		},
	})
}

func TestAccChatbotSlackChannelConfiguration_all(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	// Get the necessary environment variables to run this test.
	workspaceName, slackChannelId := getEnvironmentVariables(t)

	var slackchannelconfiguration awstypes.SlackChannelConfiguration
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_chatbot_slack_channel_configuration.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ChatbotServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSlackChannelConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSlackChannelConfigurationConfig_all(rName, workspaceName, "ERROR", slackChannelId, "terraform1", acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSlackChannelConfigurationExists(ctx, resourceName, &slackchannelconfiguration),
					resource.TestCheckResourceAttr(resourceName, "configuration_name", rName),
					resource.TestCheckResourceAttr(resourceName, "guardrail_policy_arns.#", acctest.Ct2),
					acctest.MatchResourceAttrGlobalARN(resourceName, "guardrail_policy_arns.0", "iam", regexache.MustCompile("policy/.+-[12]$")),
					acctest.MatchResourceAttrGlobalARN(resourceName, "guardrail_policy_arns.1", "iam", regexache.MustCompile("policy/.+-[12]$")),
					acctest.MatchResourceAttrGlobalARN(resourceName, "iam_role_arn", "iam", regexache.MustCompile("role/.+-1$")),
					resource.TestCheckResourceAttr(resourceName, "logging_level", "ERROR"),
					resource.TestCheckResourceAttr(resourceName, "slack_channel_id", slackChannelId),
					resource.TestCheckResourceAttr(resourceName, "slack_channel_name", "terraform1"),
					resource.TestCheckResourceAttrSet(resourceName, "slack_team_id"),
					resource.TestCheckResourceAttr(resourceName, "sns_topic_arns.#", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "user_authorization_required", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
				ExpectNonEmptyPlan: false,
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccSlackChannelConfigurationConfig_allUpdated(rName, workspaceName, "INFO", slackChannelId, "terraform2", acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSlackChannelConfigurationExists(ctx, resourceName, &slackchannelconfiguration),
					resource.TestCheckResourceAttr(resourceName, "configuration_name", rName),
					resource.TestCheckResourceAttr(resourceName, "guardrail_policy_arns.#", acctest.Ct1),
					acctest.MatchResourceAttrGlobalARN(resourceName, "guardrail_policy_arns.0", "iam", regexache.MustCompile("policy/.+-1$")),
					acctest.MatchResourceAttrGlobalARN(resourceName, "iam_role_arn", "iam", regexache.MustCompile("role/.+-2$")),
					resource.TestCheckResourceAttr(resourceName, "logging_level", "INFO"),
					resource.TestCheckResourceAttr(resourceName, "slack_channel_id", slackChannelId),
					resource.TestCheckResourceAttr(resourceName, "slack_channel_name", "terraform2"),
					resource.TestCheckResourceAttrSet(resourceName, "slack_team_id"),
					resource.TestCheckResourceAttr(resourceName, "sns_topic_arns.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "user_authorization_required", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
				ExpectNonEmptyPlan: false,
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccChatbotSlackChannelConfiguration_guardrailPolicyArns(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	// Get the necessary environment variables to run this test.
	workspaceName, slackChannelId := getEnvironmentVariables(t)

	var slackchannelconfiguration awstypes.SlackChannelConfiguration
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_chatbot_slack_channel_configuration.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ChatbotServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSlackChannelConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSlackChannelConfigurationConfig_guardrailPolicyArns(rName, workspaceName, slackChannelId),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSlackChannelConfigurationExists(ctx, resourceName, &slackchannelconfiguration),
					resource.TestCheckResourceAttr(resourceName, "configuration_name", rName),
					resource.TestCheckResourceAttr(resourceName, "guardrail_policy_arns.#", acctest.Ct1),
					acctest.MatchResourceAttrGlobalARN(resourceName, "guardrail_policy_arns.0", "iam", regexache.MustCompile("policy/.+-1$")),
					acctest.MatchResourceAttrGlobalARN(resourceName, "iam_role_arn", "iam", regexache.MustCompile("role/.+-1$")),
					resource.TestCheckResourceAttr(resourceName, "slack_channel_id", slackChannelId),
					resource.TestCheckResourceAttrSet(resourceName, "slack_team_id"),
				),
				ExpectNonEmptyPlan: false,
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"slack_channel_name", // This attribute is not always returned by the Create API immediately. And sometimes never. See the comment in TestAccChatbotSlackChannelConfiguration_basic
				},
			},
			{
				Config: testAccSlackChannelConfigurationConfig_guardrailPolicyArnsUpdated(rName, workspaceName, slackChannelId),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSlackChannelConfigurationExists(ctx, resourceName, &slackchannelconfiguration),
					resource.TestCheckResourceAttr(resourceName, "configuration_name", rName),
					resource.TestCheckResourceAttr(resourceName, "guardrail_policy_arns.#", acctest.Ct2),
					acctest.MatchResourceAttrGlobalARN(resourceName, "guardrail_policy_arns.0", "iam", regexache.MustCompile("policy/.+-1$")),
					acctest.MatchResourceAttrGlobalARN(resourceName, "guardrail_policy_arns.1", "iam", regexache.MustCompile("policy/.+-2$")),
					acctest.MatchResourceAttrGlobalARN(resourceName, "iam_role_arn", "iam", regexache.MustCompile("role/.+-1$")),
					resource.TestCheckResourceAttr(resourceName, "slack_channel_id", slackChannelId),
					resource.TestCheckResourceAttrSet(resourceName, "slack_team_id"),
				),
				ExpectNonEmptyPlan: false,
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"slack_channel_name", // This attribute is not always returned by the Create API immediately. And sometimes never. See the comment in TestAccChatbotSlackChannelConfiguration_basic
				},
			},
		},
	})
}

func TestAccChatbotSlackChannelConfiguration_iamRoleArn(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	// Get the necessary environment variables to run this test.
	workspaceName, slackChannelId := getEnvironmentVariables(t)

	var slackchannelconfiguration awstypes.SlackChannelConfiguration
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_chatbot_slack_channel_configuration.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ChatbotServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSlackChannelConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSlackChannelConfigurationConfig_iamRoleArn(rName, workspaceName, slackChannelId),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSlackChannelConfigurationExists(ctx, resourceName, &slackchannelconfiguration),
					resource.TestCheckResourceAttr(resourceName, "configuration_name", rName),
					acctest.MatchResourceAttrGlobalARN(resourceName, "iam_role_arn", "iam", regexache.MustCompile("role/.+-1$")),
					resource.TestCheckResourceAttr(resourceName, "slack_channel_id", slackChannelId),
					resource.TestCheckResourceAttrSet(resourceName, "slack_team_id"),
				),
				ExpectNonEmptyPlan: false,
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"slack_channel_name", // This attribute is not always returned by the Create API immediately. And sometimes never. See the comment in TestAccChatbotSlackChannelConfiguration_basic
				},
			},
			{
				Config: testAccSlackChannelConfigurationConfig_iamRoleArnUpdated(rName, workspaceName, slackChannelId),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSlackChannelConfigurationExists(ctx, resourceName, &slackchannelconfiguration),
					resource.TestCheckResourceAttr(resourceName, "configuration_name", rName),
					acctest.MatchResourceAttrGlobalARN(resourceName, "iam_role_arn", "iam", regexache.MustCompile("role/.+-2$")),
					resource.TestCheckResourceAttr(resourceName, "slack_channel_id", slackChannelId),
					resource.TestCheckResourceAttrSet(resourceName, "slack_team_id"),
				),
				ExpectNonEmptyPlan: false,
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"slack_channel_name", // This attribute is not always returned by the Create API immediately. And sometimes never. See the comment in TestAccChatbotSlackChannelConfiguration_basic
				},
			},
		},
	})
}

func TestAccChatbotSlackChannelConfiguration_loggingLevel(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	// Get the necessary environment variables to run this test.
	workspaceName, slackChannelId := getEnvironmentVariables(t)

	var slackchannelconfiguration awstypes.SlackChannelConfiguration
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_chatbot_slack_channel_configuration.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ChatbotServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSlackChannelConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSlackChannelConfigurationConfig_loggingLevel(rName, workspaceName, slackChannelId, "INFO"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSlackChannelConfigurationExists(ctx, resourceName, &slackchannelconfiguration),
					resource.TestCheckResourceAttr(resourceName, "configuration_name", rName),
					acctest.MatchResourceAttrGlobalARN(resourceName, "iam_role_arn", "iam", regexache.MustCompile("role/.+-1$")),
					resource.TestCheckResourceAttr(resourceName, "logging_level", "INFO"),
					resource.TestCheckResourceAttr(resourceName, "slack_channel_id", slackChannelId),
					resource.TestCheckResourceAttrSet(resourceName, "slack_team_id"),
				),
				ExpectNonEmptyPlan: false,
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"slack_channel_name", // This attribute is not always returned by the Create API immediately. And sometimes never. See the comment in TestAccChatbotSlackChannelConfiguration_basic
				},
			},
			{
				Config: testAccSlackChannelConfigurationConfig_loggingLevel(rName, workspaceName, slackChannelId, "ERROR"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSlackChannelConfigurationExists(ctx, resourceName, &slackchannelconfiguration),
					resource.TestCheckResourceAttr(resourceName, "configuration_name", rName),
					acctest.MatchResourceAttrGlobalARN(resourceName, "iam_role_arn", "iam", regexache.MustCompile("role/.+-1$")),
					resource.TestCheckResourceAttr(resourceName, "logging_level", "ERROR"),
					resource.TestCheckResourceAttr(resourceName, "slack_channel_id", slackChannelId),
					resource.TestCheckResourceAttrSet(resourceName, "slack_team_id"),
				),
				ExpectNonEmptyPlan: false,
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"slack_channel_name", // This attribute is not always returned by the Create API immediately. And sometimes never. See the comment in TestAccChatbotSlackChannelConfiguration_basic
				},
			},
		},
	})
}

func TestAccChatbotSlackChannelConfiguration_slackChannelName(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	// Get the necessary environment variables to run this test.
	workspaceName, slackChannelId := getEnvironmentVariables(t)

	var slackchannelconfiguration awstypes.SlackChannelConfiguration
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_chatbot_slack_channel_configuration.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ChatbotServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSlackChannelConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSlackChannelConfigurationConfig_slackChannelName(rName, workspaceName, slackChannelId, "test_chatbot_slack_channel_1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSlackChannelConfigurationExists(ctx, resourceName, &slackchannelconfiguration),
					resource.TestCheckResourceAttr(resourceName, "configuration_name", rName),
					acctest.MatchResourceAttrGlobalARN(resourceName, "iam_role_arn", "iam", regexache.MustCompile("role/.+-1$")),
					resource.TestCheckResourceAttr(resourceName, "slack_channel_name", "test_chatbot_slack_channel_1"),
					resource.TestCheckResourceAttr(resourceName, "slack_channel_id", slackChannelId),
					resource.TestCheckResourceAttrSet(resourceName, "slack_channel_name"),
					resource.TestCheckResourceAttrSet(resourceName, "slack_team_id"),
				),
				ExpectNonEmptyPlan: false,
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccSlackChannelConfigurationConfig_slackChannelName(rName, workspaceName, slackChannelId, "test_chatbot_slack_channel_2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSlackChannelConfigurationExists(ctx, resourceName, &slackchannelconfiguration),
					resource.TestCheckResourceAttr(resourceName, "configuration_name", rName),
					acctest.MatchResourceAttrGlobalARN(resourceName, "iam_role_arn", "iam", regexache.MustCompile("role/.+-1$")),
					resource.TestCheckResourceAttr(resourceName, "slack_channel_name", "test_chatbot_slack_channel_2"),
					resource.TestCheckResourceAttr(resourceName, "slack_channel_id", slackChannelId),
					resource.TestCheckResourceAttrSet(resourceName, "slack_channel_name"),
					resource.TestCheckResourceAttrSet(resourceName, "slack_team_id"),
				),
				ExpectNonEmptyPlan: false,
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccChatbotSlackChannelConfiguration_userAuthorizationRequired(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	// Get the necessary environment variables to run this test.
	workspaceName, slackChannelId := getEnvironmentVariables(t)

	var slackchannelconfiguration awstypes.SlackChannelConfiguration
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_chatbot_slack_channel_configuration.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ChatbotServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSlackChannelConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSlackChannelConfigurationConfig_userAuthorizationRequired(rName, workspaceName, slackChannelId, "true"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSlackChannelConfigurationExists(ctx, resourceName, &slackchannelconfiguration),
					resource.TestCheckResourceAttr(resourceName, "configuration_name", rName),
					acctest.MatchResourceAttrGlobalARN(resourceName, "iam_role_arn", "iam", regexache.MustCompile("role/.+-1$")),
					resource.TestCheckResourceAttr(resourceName, "user_authorization_required", "true"),
					resource.TestCheckResourceAttr(resourceName, "slack_channel_id", slackChannelId),
					resource.TestCheckResourceAttrSet(resourceName, "slack_team_id"),
				),
				ExpectNonEmptyPlan: false,
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"slack_channel_name", // This attribute is not always returned by the Create API immediately. And sometimes never. See the comment in TestAccChatbotSlackChannelConfiguration_basic
				},
			},
			{
				Config: testAccSlackChannelConfigurationConfig_userAuthorizationRequired(rName, workspaceName, slackChannelId, "false"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSlackChannelConfigurationExists(ctx, resourceName, &slackchannelconfiguration),
					resource.TestCheckResourceAttr(resourceName, "configuration_name", rName),
					acctest.MatchResourceAttrGlobalARN(resourceName, "iam_role_arn", "iam", regexache.MustCompile("role/.+-1$")),
					resource.TestCheckResourceAttr(resourceName, "user_authorization_required", "false"),
					resource.TestCheckResourceAttr(resourceName, "slack_channel_id", slackChannelId),
					resource.TestCheckResourceAttrSet(resourceName, "slack_team_id"),
				),
				ExpectNonEmptyPlan: false,
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"slack_channel_name", // This attribute is not always returned by the Create API immediately. And sometimes never. See the comment in TestAccChatbotSlackChannelConfiguration_basic
				},
			},
		},
	})
}

func TestAccChatbotSlackChannelConfiguration_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	// Get the necessary environment variables to run this test.
	workspaceName, slackChannelId := getEnvironmentVariables(t)

	var slackchannelconfiguration awstypes.SlackChannelConfiguration
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_chatbot_slack_channel_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			//acctest.PreCheckPartitionHasService(t, names.Chatbot)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ChatbotServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSlackChannelConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSlackChannelConfigurationConfig_basic(rName, workspaceName, slackChannelId),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSlackChannelConfigurationExists(ctx, resourceName, &slackchannelconfiguration),
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tfchatbot.ResourceSlackChannelConfiguration, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccChatbotSlackChannelConfiguration_snsTopicArns(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	// Get the necessary environment variables to run this test.
	workspaceName, slackChannelId := getEnvironmentVariables(t)

	var slackchannelconfiguration awstypes.SlackChannelConfiguration
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_chatbot_slack_channel_configuration.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ChatbotServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSlackChannelConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSlackChannelConfigurationConfig_snsTopicArns(rName, workspaceName, slackChannelId),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSlackChannelConfigurationExists(ctx, resourceName, &slackchannelconfiguration),
					resource.TestCheckResourceAttr(resourceName, "configuration_name", rName),
					resource.TestCheckResourceAttr(resourceName, "sns_topic_arns.#", acctest.Ct1),
					acctest.MatchResourceAttrGlobalARN(resourceName, "iam_role_arn", "iam", regexache.MustCompile("role/.+-1$")),
					resource.TestCheckResourceAttr(resourceName, "slack_channel_id", slackChannelId),
					resource.TestCheckResourceAttrSet(resourceName, "slack_team_id"),
				),
				ExpectNonEmptyPlan: false,
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"slack_channel_name", // This attribute is not always returned by the Create API immediately. And sometimes never. See the comment in TestAccChatbotSlackChannelConfiguration_basic
				},
			},
			{
				Config: testAccSlackChannelConfigurationConfig_snsTopicArnsUpdated(rName, workspaceName, slackChannelId),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSlackChannelConfigurationExists(ctx, resourceName, &slackchannelconfiguration),
					resource.TestCheckResourceAttr(resourceName, "configuration_name", rName),
					resource.TestCheckResourceAttr(resourceName, "sns_topic_arns.#", acctest.Ct2),
					acctest.MatchResourceAttrGlobalARN(resourceName, "iam_role_arn", "iam", regexache.MustCompile("role/.+-1$")),
					resource.TestCheckResourceAttr(resourceName, "slack_channel_id", slackChannelId),
					resource.TestCheckResourceAttrSet(resourceName, "slack_team_id"),
				),
				ExpectNonEmptyPlan: false,
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"slack_channel_name", // This attribute is not always returned by the Create API immediately. And sometimes never. See the comment in TestAccChatbotSlackChannelConfiguration_basic
				},
			},
		},
	})
}

func TestAccChatbotSlackChannelConfiguration_tags(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	// Get the necessary environment variables to run this test.
	workspaceName, slackChannelId := getEnvironmentVariables(t)

	var slackchannelconfiguration awstypes.SlackChannelConfiguration
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_chatbot_slack_channel_configuration.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SecurityLakeServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSlackChannelConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccChatbotSlackChannelConfiguration_tags1(rName, workspaceName, slackChannelId, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSlackChannelConfigurationExists(ctx, resourceName, &slackchannelconfiguration),
					resource.TestCheckResourceAttr(resourceName, "configuration_name", rName),
					acctest.MatchResourceAttrGlobalARN(resourceName, "iam_role_arn", "iam", regexache.MustCompile("role/.+-1$")),
					resource.TestCheckResourceAttr(resourceName, "slack_channel_id", slackChannelId),
					resource.TestCheckResourceAttrSet(resourceName, "slack_team_id"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"slack_channel_name", // This attribute is not always returned by the Create API immediately. And sometimes never. See the comment in TestAccChatbotSlackChannelConfiguration_basic
				},
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

			_, err := tfchatbot.FindSlackChannelConfigurationByID(ctx, conn, *aws.String(rs.Primary.Attributes["chat_configuration_arn"]))
			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return create.Error(names.Chatbot, create.ErrActionCheckingDestroyed, tfchatbot.ResNameSlackChannelConfiguration, rs.Primary.ID, err)
			}

			return create.Error(names.Chatbot, create.ErrActionCheckingDestroyed, tfchatbot.ResNameSlackChannelConfiguration, rs.Primary.ID, errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccCheckSlackChannelConfigurationExists(ctx context.Context, name string, slackchannelconfiguration *awstypes.SlackChannelConfiguration) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.Chatbot, create.ErrActionCheckingExistence, tfchatbot.ResNameSlackChannelConfiguration, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.Chatbot, create.ErrActionCheckingExistence, tfchatbot.ResNameSlackChannelConfiguration, name, errors.New("not set"))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).ChatbotClient(ctx)
		resp, err := tfchatbot.FindSlackChannelConfigurationByID(ctx, conn, *aws.String(rs.Primary.Attributes["chat_configuration_arn"]))

		if err != nil {
			return create.Error(names.Chatbot, create.ErrActionCheckingExistence, tfchatbot.ResNameSlackChannelConfiguration, rs.Primary.ID, err)
		}

		*slackchannelconfiguration = *resp

		return nil
	}
}

const testAccSlackChannelConfigurationConfig_base = `

// Test IAM policy
resource "aws_iam_policy" "test_guardrail_policy_1" {
  name        = "%[1]s-1"
  path        = "/"
  description = "Guardrail policy to deny EC2 actions"

  policy = <<EOF
{
  "Statement": [
    {
      "Action": [
        "ec2:*"
      ],
      "Effect": "Deny",
      "Resource": "*"
    }
  ],
  "Version": "2012-10-17"
}
EOF
}

resource "aws_iam_policy" "test_guardrail_policy_2" {
  name        = "%[1]s-2"
  path        = "/"
  description = "Guardrail policy to deny EC2 actions"

  policy = <<EOF
{
  "Statement": [
    {
      "Action": [
        "ec2:*"
      ],
      "Effect": "Deny",
      "Resource": "*"
    }
  ],
  "Version": "2012-10-17"
}
EOF
}

resource "aws_sns_topic" "test_sns_topic_1" {
  name = "%[1]s-1"
}

resource "aws_sns_topic" "test_sns_topic_2" {
  name = "%[1]s-2"
}

resource "aws_iam_role" "test_chatbot_role_1" {
  name   = "%[1]s-1"

  assume_role_policy = <<EOF
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Effect": "Allow",
            "Principal": {
                "Service": "chatbot.amazonaws.com"
            },
            "Action": "sts:AssumeRole"
        }
    ]
}
EOF
}

resource "aws_iam_role" "test_chatbot_role_2" {
  name   = "%[1]s-2"

  assume_role_policy = <<EOF
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Effect": "Allow",
            "Principal": {
                "Service": "chatbot.amazonaws.com"
            },
            "Action": "sts:AssumeRole"
        }
    ]
}
EOF
}

data "aws_chatbot_slack_workspace" "test" {
  slack_team_name = "%[2]s"
}
`

func testAccSlackChannelConfigurationConfig_basic(rName string, workspaceName string, slackChannelId string) string {
	return acctest.ConfigCompose(fmt.Sprintf(testAccSlackChannelConfigurationConfig_base, rName, workspaceName), fmt.Sprintf(`

resource "aws_chatbot_slack_channel_configuration" "test" {
  configuration_name   = %[1]q
  iam_role_arn         = aws_iam_role.test_chatbot_role_1.arn
  slack_channel_id     = "%[2]s"
  slack_team_id        = data.aws_chatbot_slack_workspace.test.slack_team_id
}
`, rName, slackChannelId))
}

func testAccSlackChannelConfigurationConfig_all(rName string, workspaceName string, loggingLevel string, slackChannelId string, slackChannelName string, tagKey1 string, tagValue1 string) string {
	return acctest.ConfigCompose(fmt.Sprintf(testAccSlackChannelConfigurationConfig_base, rName, workspaceName), fmt.Sprintf(`
resource "aws_chatbot_slack_channel_configuration" "test" {
  configuration_name   = %[1]q
  guardrail_policy_arns   = [aws_iam_policy.test_guardrail_policy_1.arn, aws_iam_policy.test_guardrail_policy_2.arn]
  iam_role_arn         = aws_iam_role.test_chatbot_role_1.arn
  logging_level        = "%[2]s"
  slack_channel_id     = "%[3]s"
  slack_channel_name   = "%[4]s"
  slack_team_id        = data.aws_chatbot_slack_workspace.test.slack_team_id
  sns_topic_arns       = [aws_sns_topic.test_sns_topic_1.arn, aws_sns_topic.test_sns_topic_2.arn]
  user_authorization_required = true
  tags = {
    %[5]q = %[6]q
   }
}
`, rName, loggingLevel, slackChannelId, slackChannelName, tagKey1, tagValue1))
}

func testAccSlackChannelConfigurationConfig_allUpdated(rName string, workspaceName string, loggingLevel string, slackChannelId string, slackChannelName string, tagKey1 string, tagValue1 string) string {
	return acctest.ConfigCompose(fmt.Sprintf(testAccSlackChannelConfigurationConfig_base, rName, workspaceName), fmt.Sprintf(`
resource "aws_chatbot_slack_channel_configuration" "test" {
  configuration_name   = %[1]q
  guardrail_policy_arns   = [aws_iam_policy.test_guardrail_policy_1.arn]
  iam_role_arn         = aws_iam_role.test_chatbot_role_2.arn
  logging_level        = "%[2]s"
  slack_channel_id     = "%[3]s"
  slack_channel_name   = "%[4]s"
  slack_team_id        = data.aws_chatbot_slack_workspace.test.slack_team_id
  sns_topic_arns       = [aws_sns_topic.test_sns_topic_2.arn]
  user_authorization_required = false
  tags = {
    %[5]q = %[6]q
   }
}
`, rName, loggingLevel, slackChannelId, slackChannelName, tagKey1, tagValue1))
}

func testAccSlackChannelConfigurationConfig_guardrailPolicyArns(rName string, workspaceName string, slackChannelId string) string {
	return acctest.ConfigCompose(fmt.Sprintf(testAccSlackChannelConfigurationConfig_base, rName, workspaceName), fmt.Sprintf(`
resource "aws_chatbot_slack_channel_configuration" "test" {
  configuration_name   = %[1]q
  guardrail_policy_arns   = [aws_iam_policy.test_guardrail_policy_1.arn]
  iam_role_arn         = aws_iam_role.test_chatbot_role_1.arn
  slack_channel_id     = "%[2]s"
  slack_team_id        = data.aws_chatbot_slack_workspace.test.slack_team_id
}
`, rName, slackChannelId))
}

func testAccSlackChannelConfigurationConfig_guardrailPolicyArnsUpdated(rName string, workspaceName string, slackChannelId string) string {
	return acctest.ConfigCompose(fmt.Sprintf(testAccSlackChannelConfigurationConfig_base, rName, workspaceName), fmt.Sprintf(`
resource "aws_chatbot_slack_channel_configuration" "test" {
  configuration_name   = %[1]q
  guardrail_policy_arns   = [aws_iam_policy.test_guardrail_policy_1.arn, aws_iam_policy.test_guardrail_policy_2.arn]
  iam_role_arn         = aws_iam_role.test_chatbot_role_1.arn
  slack_channel_id     = "%[2]s"
  slack_team_id        = data.aws_chatbot_slack_workspace.test.slack_team_id
}
`, rName, slackChannelId))
}

func testAccSlackChannelConfigurationConfig_iamRoleArn(rName string, workspaceName string, slackChannelId string) string {
	return acctest.ConfigCompose(fmt.Sprintf(testAccSlackChannelConfigurationConfig_base, rName, workspaceName), fmt.Sprintf(`
resource "aws_chatbot_slack_channel_configuration" "test" {
  configuration_name   = %[1]q
  iam_role_arn         = aws_iam_role.test_chatbot_role_1.arn
  slack_channel_id     = "%[2]s"
  slack_team_id        = data.aws_chatbot_slack_workspace.test.slack_team_id
}
`, rName, slackChannelId))
}

func testAccSlackChannelConfigurationConfig_iamRoleArnUpdated(rName string, workspaceName string, slackChannelId string) string {
	return acctest.ConfigCompose(fmt.Sprintf(testAccSlackChannelConfigurationConfig_base, rName, workspaceName), fmt.Sprintf(`
resource "aws_chatbot_slack_channel_configuration" "test" {
  configuration_name   = %[1]q
  iam_role_arn         = aws_iam_role.test_chatbot_role_2.arn
  slack_channel_id     = "%[2]s"
  slack_team_id        = data.aws_chatbot_slack_workspace.test.slack_team_id
}
`, rName, slackChannelId))
}

func testAccSlackChannelConfigurationConfig_loggingLevel(rName string, workspaceName string, slackChannelId string, loggingLevel string) string {
	return acctest.ConfigCompose(fmt.Sprintf(testAccSlackChannelConfigurationConfig_base, rName, workspaceName), fmt.Sprintf(`
resource "aws_chatbot_slack_channel_configuration" "test" {
  configuration_name   = %[1]q
  iam_role_arn         = aws_iam_role.test_chatbot_role_1.arn
  logging_level        = "%[3]s"
  slack_channel_id     = "%[2]s"
  slack_team_id        = data.aws_chatbot_slack_workspace.test.slack_team_id
}
`, rName, slackChannelId, loggingLevel))
}

func testAccSlackChannelConfigurationConfig_slackChannelName(rName string, workspaceName string, slackChannelId string, slackChannelName string) string {
	return acctest.ConfigCompose(fmt.Sprintf(testAccSlackChannelConfigurationConfig_base, rName, workspaceName), fmt.Sprintf(`
resource "aws_chatbot_slack_channel_configuration" "test" {
  configuration_name   = %[1]q
  iam_role_arn         = aws_iam_role.test_chatbot_role_1.arn
  slack_channel_id     = "%[2]s"
  slack_team_id        = data.aws_chatbot_slack_workspace.test.slack_team_id
  slack_channel_name   = "%[3]s"
}
`, rName, slackChannelId, slackChannelName))
}

func testAccSlackChannelConfigurationConfig_snsTopicArns(rName string, workspaceName string, slackChannelId string) string {
	return acctest.ConfigCompose(fmt.Sprintf(testAccSlackChannelConfigurationConfig_base, rName, workspaceName), fmt.Sprintf(`
resource "aws_chatbot_slack_channel_configuration" "test" {
  configuration_name   = %[1]q
  sns_topic_arns       = [aws_sns_topic.test_sns_topic_1.arn]
  iam_role_arn         = aws_iam_role.test_chatbot_role_1.arn
  slack_channel_id     = "%[2]s"
  slack_team_id        = data.aws_chatbot_slack_workspace.test.slack_team_id
}
`, rName, slackChannelId))
}

func testAccSlackChannelConfigurationConfig_snsTopicArnsUpdated(rName string, workspaceName string, slackChannelId string) string {
	return acctest.ConfigCompose(fmt.Sprintf(testAccSlackChannelConfigurationConfig_base, rName, workspaceName), fmt.Sprintf(`
resource "aws_chatbot_slack_channel_configuration" "test" {
  configuration_name   = %[1]q
  sns_topic_arns       = [aws_sns_topic.test_sns_topic_1.arn, aws_sns_topic.test_sns_topic_2.arn]
  iam_role_arn         = aws_iam_role.test_chatbot_role_1.arn
  slack_channel_id     = "%[2]s"
  slack_team_id        = data.aws_chatbot_slack_workspace.test.slack_team_id
}
`, rName, slackChannelId))
}

func testAccSlackChannelConfigurationConfig_userAuthorizationRequired(rName string, workspaceName string, slackChannelId string, userAuthorizationRequired string) string {
	return acctest.ConfigCompose(fmt.Sprintf(testAccSlackChannelConfigurationConfig_base, rName, workspaceName), fmt.Sprintf(`
resource "aws_chatbot_slack_channel_configuration" "test" {
  configuration_name   = %[1]q
  iam_role_arn         = aws_iam_role.test_chatbot_role_1.arn
  slack_channel_id     = "%[2]s"
  slack_team_id        = data.aws_chatbot_slack_workspace.test.slack_team_id
  user_authorization_required = %[3]s
}
`, rName, slackChannelId, userAuthorizationRequired))
}

func testAccChatbotSlackChannelConfiguration_tags1(rName string, workspaceName string, slackChannelId string, tag1Key, tag1Value string) string {
	return acctest.ConfigCompose(fmt.Sprintf(testAccSlackChannelConfigurationConfig_base, rName, workspaceName), fmt.Sprintf(`
resource "aws_chatbot_slack_channel_configuration" "test" {
  configuration_name   = %[1]q
  iam_role_arn         = aws_iam_role.test_chatbot_role_1.arn
  slack_channel_id     = "%[2]s"
  slack_team_id        = data.aws_chatbot_slack_workspace.test.slack_team_id
  tags = {
    %[3]q = %[4]q
  }
}
`, rName, slackChannelId, tag1Key, tag1Value))
}

func testAccChatbotSlackChannelConfiguration_tags2(rName string, workspaceName string, slackChannelId string, tag1Key string, tag1Value string, tag2Key string, tag2Value string) string {
	return acctest.ConfigCompose(fmt.Sprintf(testAccSlackChannelConfigurationConfig_base, rName, workspaceName), fmt.Sprintf(`
resource "aws_chatbot_slack_channel_configuration" "test" {
  configuration_name   = %[1]q
  iam_role_arn         = aws_iam_role.test_chatbot_role_1.arn
  slack_channel_id     = "%[2]s"
  slack_team_id        = data.aws_chatbot_slack_workspace.test.slack_team_id
  tags = {
    %[3]q = %[4]q
	%[5]q = %[6]q

  }
}
`, rName, slackChannelId, tag1Key, tag1Value, tag2Key, tag2Value))
}
