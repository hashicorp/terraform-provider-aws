// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package chatbot_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
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
	testResourceTeamsChannelConfiguration = "aws_chatbot_teams_channel_configuration.test"

	envTeamsChannelID = "CHATBOT_TEAMS_CHANNEL_ID"
	envTeamsTeamID    = "CHATBOT_TEAMS_TEAM_ID"
	envTeamsTenantID  = "CHATBOT_TEAMS_TENANT_ID"
)

func TestAccChatbotTeamsChannelConfiguration_serial(t *testing.T) {
	t.Parallel()

	testCases := map[string]func(t *testing.T){
		acctest.CtBasic:      testAccTeamsChannelConfiguration_basic,
		acctest.CtDisappears: testAccTeamsChannelConfiguration_disappears,
	}

	acctest.RunSerialTests1Level(t, testCases, 0)
}

func testAccTeamsChannelConfiguration_basic(t *testing.T) {
	ctx := acctest.Context(t)

	var teamschannelconfiguration types.TeamsChannelConfiguration
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	// The teams workspace must be created via the AWS Console. It cannot be created via APIs or Terraform.
	// Once it is created, export the name of the workspace in the env variable for this test
	teamID := acctest.SkipIfEnvVarNotSet(t, envTeamsTeamID)
	channelID := acctest.SkipIfEnvVarNotSet(t, envTeamsChannelID)
	tenantID := acctest.SkipIfEnvVarNotSet(t, envTeamsTenantID)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ChatbotServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTeamsChannelConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTeamsChannelConfigurationConfig_basic(rName, channelID, teamID, tenantID),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTeamsChannelConfigurationExists(ctx, testResourceTeamsChannelConfiguration, &teamschannelconfiguration),
					resource.TestCheckResourceAttr(testResourceTeamsChannelConfiguration, "configuration_name", rName),
					acctest.MatchResourceAttrGlobalARN(ctx, testResourceTeamsChannelConfiguration, "chat_configuration_arn", "chatbot", regexache.MustCompile(fmt.Sprintf(`chat-configuration/.*/%s`, rName))),
					resource.TestCheckResourceAttrPair(testResourceTeamsChannelConfiguration, names.AttrIAMRoleARN, "aws_iam_role.test", names.AttrARN),
					resource.TestCheckResourceAttr(testResourceTeamsChannelConfiguration, "channel_id", channelID),
					resource.TestCheckResourceAttrSet(testResourceTeamsChannelConfiguration, "channel_name"),
					resource.TestCheckResourceAttr(testResourceTeamsChannelConfiguration, "team_id", teamID),
					resource.TestCheckResourceAttrSet(testResourceTeamsChannelConfiguration, "team_name"),
				),
			},
			{
				ResourceName:                         testResourceTeamsChannelConfiguration,
				ImportState:                          true,
				ImportStateIdFunc:                    testAccTeamsChannelConfigurationImportStateIDFunc(testResourceTeamsChannelConfiguration),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "team_id",
			},
		},
	})
}

func testAccTeamsChannelConfiguration_disappears(t *testing.T) {
	ctx := acctest.Context(t)

	var teamschannelconfiguration types.TeamsChannelConfiguration
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	// The teams workspace must be created via the AWS Console. It cannot be created via APIs or Terraform.
	// Once it is created, export the name of the workspace in the env variable for this test
	teamID := acctest.SkipIfEnvVarNotSet(t, envTeamsTeamID)
	channelID := acctest.SkipIfEnvVarNotSet(t, envTeamsChannelID)
	tenantID := acctest.SkipIfEnvVarNotSet(t, envTeamsTenantID)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ChatbotServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTeamsChannelConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTeamsChannelConfigurationConfig_basic(rName, channelID, teamID, tenantID),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTeamsChannelConfigurationExists(ctx, testResourceTeamsChannelConfiguration, &teamschannelconfiguration),
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tfchatbot.ResourceTeamsChannelConfiguration, testResourceTeamsChannelConfiguration),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckTeamsChannelConfigurationDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).ChatbotClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_chatbot_teams_channel_configuration" {
				continue
			}

			_, err := tfchatbot.FindTeamsChannelConfigurationByTeamID(ctx, conn, rs.Primary.Attributes["team_id"])

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return create.Error(names.Chatbot, create.ErrActionCheckingDestroyed, tfchatbot.ResNameTeamsChannelConfiguration, rs.Primary.Attributes["team_id"], errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccCheckTeamsChannelConfigurationExists(ctx context.Context, name string, teamschannelconfiguration *types.TeamsChannelConfiguration) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.Chatbot, create.ErrActionCheckingExistence, tfchatbot.ResNameTeamsChannelConfiguration, name, errors.New("not found"))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).ChatbotClient(ctx)

		output, err := tfchatbot.FindTeamsChannelConfigurationByTeamID(ctx, conn, rs.Primary.Attributes["team_id"])

		if err != nil {
			return err
		}

		*teamschannelconfiguration = *output

		return nil
	}
}

func testAccTeamsChannelConfigurationImportStateIDFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("Not found: %s", resourceName)
		}

		return rs.Primary.Attributes["team_id"], nil
	}
}

func testAccTeamsChannelConfigurationConfig_basic(rName, channelID, teamID, tenantID string) string {
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

resource "aws_chatbot_teams_channel_configuration" "test" {
  channel_id         = %[2]q
  configuration_name = %[1]q
  iam_role_arn       = aws_iam_role.test.arn
  team_id            = %[3]q
  tenant_id          = %[4]q

  tags = {
    Name = %[2]q
  }
}
`, rName, channelID, teamID, tenantID)
}
