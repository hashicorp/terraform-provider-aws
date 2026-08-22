// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package chatbot_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/service/chatbot/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfchatbot "github.com/hashicorp/terraform-provider-aws/internal/service/chatbot"
	"github.com/hashicorp/terraform-provider-aws/names"
)

const (
	testResourceTeamsChannelConfiguration  = "aws_chatbot_teams_channel_configuration.test"
	testResourceTeamsChannelConfiguration2 = "aws_chatbot_teams_channel_configuration.test2"

	envTeamsChannelID  = "CHATBOT_TEAMS_CHANNEL_ID"
	envTeamsChannelID2 = "CHATBOT_TEAMS_CHANNEL_ID_2"
	envTeamsTeamID     = "CHATBOT_TEAMS_TEAM_ID"
	envTeamsTenantID   = "CHATBOT_TEAMS_TENANT_ID"
)

func TestAccChatbotTeamsChannelConfiguration_serial(t *testing.T) {
	t.Parallel()

	testCases := map[string]func(t *testing.T){
		acctest.CtBasic:      testAccTeamsChannelConfiguration_basic,
		acctest.CtDisappears: testAccTeamsChannelConfiguration_disappears,
		"multiple":           testAccTeamsChannelConfiguration_multiple,
	}

	acctest.RunSerialTests1Level(t, testCases, 0)
}

func testAccTeamsChannelConfiguration_basic(t *testing.T) {
	ctx := acctest.Context(t)

	var teamschannelconfiguration types.TeamsChannelConfiguration
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	// The teams workspace must be created via the AWS Console. It cannot be created via APIs or Terraform.
	// Once it is created, export the name of the workspace in the env variable for this test
	teamID := acctest.SkipIfEnvVarNotSet(t, envTeamsTeamID)
	channelID := acctest.SkipIfEnvVarNotSet(t, envTeamsChannelID)
	tenantID := acctest.SkipIfEnvVarNotSet(t, envTeamsTenantID)

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ChatbotServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTeamsChannelConfigurationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccTeamsChannelConfigurationConfig_basic(rName, channelID, teamID, tenantID),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTeamsChannelConfigurationExists(ctx, t, testResourceTeamsChannelConfiguration, &teamschannelconfiguration),
					resource.TestCheckResourceAttr(testResourceTeamsChannelConfiguration, "configuration_name", rName),
					acctest.MatchResourceAttrGlobalARN(ctx, testResourceTeamsChannelConfiguration, "chat_configuration_arn", "chatbot", regexache.MustCompile(fmt.Sprintf(`chat-configuration/.*/%s`, rName))),
					resource.TestCheckResourceAttrPair(testResourceTeamsChannelConfiguration, names.AttrIAMRoleARN, "aws_iam_role.test", names.AttrARN),
					resource.TestCheckResourceAttr(testResourceTeamsChannelConfiguration, "channel_id", channelID),
					resource.TestCheckResourceAttr(testResourceTeamsChannelConfiguration, "channel_name", rName),
					resource.TestCheckResourceAttr(testResourceTeamsChannelConfiguration, "team_id", teamID),
					resource.TestCheckResourceAttr(testResourceTeamsChannelConfiguration, "team_name", rName),
				),
			},
			{
				ResourceName:                         testResourceTeamsChannelConfiguration,
				ImportState:                          true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(testResourceTeamsChannelConfiguration, "chat_configuration_arn"),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "chat_configuration_arn",
			},
			{
				// Legacy import by team_id, still supported for backward compatibility.
				ResourceName:                         testResourceTeamsChannelConfiguration,
				ImportState:                          true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(testResourceTeamsChannelConfiguration, "team_id"),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "team_id",
			},
		},
	})
}

func testAccTeamsChannelConfiguration_multiple(t *testing.T) {
	ctx := acctest.Context(t)

	var v1, v2 types.TeamsChannelConfiguration
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	// The teams workspace must be created via the AWS Console. It cannot be created via APIs or Terraform.
	// Once it is created, export the name of the workspace in the env variable for this test.
	// AWS Chatbot rejects a second configuration for the same (team, channel) pair, so this test
	// needs a second channel in the same team.
	teamID := acctest.SkipIfEnvVarNotSet(t, envTeamsTeamID)
	channelID := acctest.SkipIfEnvVarNotSet(t, envTeamsChannelID)
	channelID2 := acctest.SkipIfEnvVarNotSet(t, envTeamsChannelID2)
	tenantID := acctest.SkipIfEnvVarNotSet(t, envTeamsTenantID)

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ChatbotServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTeamsChannelConfigurationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccTeamsChannelConfigurationConfig_multiple(rName, channelID, channelID2, teamID, tenantID),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTeamsChannelConfigurationExists(ctx, t, testResourceTeamsChannelConfiguration, &v1),
					testAccCheckTeamsChannelConfigurationExists(ctx, t, testResourceTeamsChannelConfiguration2, &v2),
					resource.TestCheckResourceAttr(testResourceTeamsChannelConfiguration, "team_id", teamID),
					resource.TestCheckResourceAttr(testResourceTeamsChannelConfiguration2, "team_id", teamID),
				),
			},
		},
	})
}

func testAccTeamsChannelConfiguration_disappears(t *testing.T) {
	ctx := acctest.Context(t)

	var teamschannelconfiguration types.TeamsChannelConfiguration
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	// The teams workspace must be created via the AWS Console. It cannot be created via APIs or Terraform.
	// Once it is created, export the name of the workspace in the env variable for this test
	teamID := acctest.SkipIfEnvVarNotSet(t, envTeamsTeamID)
	channelID := acctest.SkipIfEnvVarNotSet(t, envTeamsChannelID)
	tenantID := acctest.SkipIfEnvVarNotSet(t, envTeamsTenantID)

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ChatbotServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTeamsChannelConfigurationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccTeamsChannelConfigurationConfig_basic(rName, channelID, teamID, tenantID),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTeamsChannelConfigurationExists(ctx, t, testResourceTeamsChannelConfiguration, &teamschannelconfiguration),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfchatbot.ResourceTeamsChannelConfiguration, testResourceTeamsChannelConfiguration),
				),
				ExpectNonEmptyPlan: true,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(testResourceTeamsChannelConfiguration, plancheck.ResourceActionCreate),
					},
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(testResourceTeamsChannelConfiguration, plancheck.ResourceActionCreate),
					},
				},
			},
		},
	})
}

func testAccCheckTeamsChannelConfigurationDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).ChatbotClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_chatbot_teams_channel_configuration" {
				continue
			}

			_, err := tfchatbot.FindTeamsChannelConfigurationByARN(ctx, conn, rs.Primary.Attributes["chat_configuration_arn"])

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return create.Error(names.Chatbot, create.ErrActionCheckingDestroyed, tfchatbot.ResNameTeamsChannelConfiguration, rs.Primary.Attributes["chat_configuration_arn"], errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccCheckTeamsChannelConfigurationExists(ctx context.Context, t *testing.T, name string, teamschannelconfiguration *types.TeamsChannelConfiguration) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.Chatbot, create.ErrActionCheckingExistence, tfchatbot.ResNameTeamsChannelConfiguration, name, errors.New("not found"))
		}

		conn := acctest.ProviderMeta(ctx, t).ChatbotClient(ctx)

		output, err := tfchatbot.FindTeamsChannelConfigurationByARN(ctx, conn, rs.Primary.Attributes["chat_configuration_arn"])

		if err != nil {
			return err
		}

		*teamschannelconfiguration = *output

		return nil
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
  channel_name       = %[1]q
  configuration_name = %[1]q
  iam_role_arn       = aws_iam_role.test.arn
  team_id            = %[3]q
  team_name          = %[1]q
  tenant_id          = %[4]q

  tags = {
    Name = %[1]q
  }
}
`, rName, channelID, teamID, tenantID)
}

func testAccTeamsChannelConfigurationConfig_multiple(rName, channelID, channelID2, teamID, tenantID string) string {
	return acctest.ConfigCompose(testAccTeamsChannelConfigurationConfig_basic(rName, channelID, teamID, tenantID), fmt.Sprintf(`
resource "aws_chatbot_teams_channel_configuration" "test2" {
  channel_id         = %[2]q
  configuration_name = "%[1]s-2"
  iam_role_arn       = aws_iam_role.test.arn
  team_id            = %[3]q
  tenant_id          = %[4]q
}
`, rName, channelID2, teamID, tenantID))
}
