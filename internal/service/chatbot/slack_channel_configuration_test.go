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

func TestAccChatbotSlackChannelConfiguration_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var slackchannelconfiguration chatbot.DescribeSlackChannelConfigurationResponse
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_chatbot_slack_channel_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.ChatbotEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ChatbotServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSlackChannelConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSlackChannelConfigurationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSlackChannelConfigurationExists(ctx, resourceName, &slackchannelconfiguration),
					resource.TestCheckResourceAttr(resourceName, "auto_minor_version_upgrade", "false"),
					resource.TestCheckResourceAttrSet(resourceName, "maintenance_window_start_time.0.day_of_week"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "user.*", map[string]string{
						"console_access": "false",
						"groups.#":       "0",
						"username":       "Test",
						"password":       "TestTest1234",
					}),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "chatbot", regexache.MustCompile(`slackchannelconfiguration:+.`)),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"apply_immediately", "user"},
			},
		},
	})
}

func TestAccChatbotSlackChannelConfiguration_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var slackchannelconfiguration types.SlackChannelConfiguration
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_chatbot_slack_channel_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.ChatbotEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ChatbotServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSlackChannelConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSlackChannelConfigurationConfig_basic(rName, testAccSlackChannelConfigurationVersionNewer),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSlackChannelConfigurationExists(ctx, resourceName, &slackchannelconfiguration),
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tfchatbot.ResourceSlackChannelConfiguration, resourceName),
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

			_, err := tfchatbot.FindSlackChannelConfigurationByID(ctx, conn, rs.Primary.ID)

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

func testAccCheckSlackChannelConfigurationExists(ctx context.Context, name string, slackchannelconfiguration *types.SlackChannelConfiguration) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.Chatbot, create.ErrActionCheckingExistence, tfchatbot.ResNameSlackChannelConfiguration, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.Chatbot, create.ErrActionCheckingExistence, tfchatbot.ResNameSlackChannelConfiguration, name, errors.New("not set"))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).ChatbotClient(ctx)

		output, err := tfchatbot.FindSlackChannelConfigurationByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return create.Error(names.Chatbot, create.ErrActionCheckingExistence, tfchatbot.ResNameSlackChannelConfiguration, rs.Primary.ID, err)
		}

		*slackchannelconfiguration = *output

		return nil
	}
}

func testAccPreCheck(ctx context.Context, t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).ChatbotClient(ctx)

	_, err := conn.DescribeSlackChannelConfigurations(ctx, &chatbot.DescribeSlackChannelConfigurationsInput{})

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}
	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccSlackChannelConfigurationConfig_basic(rName, version string) string {
	return fmt.Sprintf(`
resource "aws_security_group" "test" {
  name = %[1]q
}

resource "aws_chatbot_slack_channel_configuration" "test" {
  slack_channel_configuration_name             = %[1]q
  engine_type             = "ActiveChatbot"
  engine_version          = %[2]q
  host_instance_type      = "chatbot.t2.micro"
  security_groups         = [aws_security_group.test.id]
  authentication_strategy = "simple"
  storage_type            = "efs"
}
`, rName, version)
}
