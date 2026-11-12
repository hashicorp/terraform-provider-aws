// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package pinpointsmsvoicev2_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/hashicorp/terraform-plugin-testing/compare"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfpinpointsmsvoicev2 "github.com/hashicorp/terraform-provider-aws/internal/service/pinpointsmsvoicev2"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// The account default is a single setting per Region, so these tests run serially.
func TestAccPinpointSMSVoiceV2AccountDefaultProtectConfiguration_serial(t *testing.T) {
	t.Parallel()

	testCases := map[string]func(*testing.T){
		acctest.CtBasic:      testAccAccountDefaultProtectConfiguration_basic,
		acctest.CtDisappears: testAccAccountDefaultProtectConfiguration_disappears,
		"update":             testAccAccountDefaultProtectConfiguration_update,
		"Identity":           testAccPinpointSMSVoiceV2AccountDefaultProtectConfiguration_identitySerial,
	}

	acctest.RunSerialTests1Level(t, testCases, 0)
}

func testAccAccountDefaultProtectConfiguration_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_pinpointsmsvoicev2_account_default_protect_configuration.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckProtectConfiguration(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.PinpointSMSVoiceV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAccountDefaultProtectConfigurationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccAccountDefaultProtectConfigurationConfig_basic(0),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAccountDefaultProtectConfigurationExists(ctx, t, resourceName),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrID), knownvalue.StringExact(acctest.Region())),
					statecheck.CompareValuePairs(resourceName, tfjsonpath.New("protect_configuration_id"), "aws_pinpointsmsvoicev2_protect_configuration.test[0]", tfjsonpath.New(names.AttrID), compare.ValuesSame()),
				},
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateId:                        acctest.Region(),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: names.AttrRegion,
			},
		},
	})
}

func testAccAccountDefaultProtectConfiguration_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_pinpointsmsvoicev2_account_default_protect_configuration.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckProtectConfiguration(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.PinpointSMSVoiceV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAccountDefaultProtectConfigurationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccAccountDefaultProtectConfigurationConfig_basic(0),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAccountDefaultProtectConfigurationExists(ctx, t, resourceName),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfpinpointsmsvoicev2.ResourceAccountDefaultProtectConfiguration, resourceName),
				),
				ExpectNonEmptyPlan: true,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
		},
	})
}

func testAccAccountDefaultProtectConfiguration_update(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_pinpointsmsvoicev2_account_default_protect_configuration.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckProtectConfiguration(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.PinpointSMSVoiceV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAccountDefaultProtectConfigurationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccAccountDefaultProtectConfigurationConfig_basic(0),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAccountDefaultProtectConfigurationExists(ctx, t, resourceName),
				),
			},
			{
				Config: testAccAccountDefaultProtectConfigurationConfig_basic(1),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAccountDefaultProtectConfigurationExists(ctx, t, resourceName),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.CompareValuePairs(resourceName, tfjsonpath.New("protect_configuration_id"), "aws_pinpointsmsvoicev2_protect_configuration.test[1]", tfjsonpath.New(names.AttrID), compare.ValuesSame()),
				},
			},
		},
	})
}

func testAccCheckAccountDefaultProtectConfigurationDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_pinpointsmsvoicev2_account_default_protect_configuration" {
				continue
			}

			ctx := acctest.NewTestResourceContext(ctx, rs.Type, rs.Primary.Attributes[names.AttrRegion])
			conn := acctest.ProviderMeta(ctx, t).PinpointSMSVoiceV2Client(ctx)
			output, err := tfpinpointsmsvoicev2.FindAccountDefaultProtectConfiguration(ctx, conn)

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("End User Messaging SMS Account Default Protect Configuration %s still exists", aws.ToString(output.ProtectConfigurationId))
		}

		return nil
	}
}

// testAccCheckAccountDefaultProtectConfigurationExists checks that AWS has the protect configuration in
// the state of n as the account default.
func testAccCheckAccountDefaultProtectConfigurationExists(ctx context.Context, t *testing.T, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		ctx := acctest.NewTestResourceContext(ctx, rs.Type, rs.Primary.Attributes[names.AttrRegion])
		conn := acctest.ProviderMeta(ctx, t).PinpointSMSVoiceV2Client(ctx)

		output, err := tfpinpointsmsvoicev2.FindAccountDefaultProtectConfiguration(ctx, conn)
		if err != nil {
			return err
		}

		if got, want := aws.ToString(output.ProtectConfigurationId), rs.Primary.Attributes["protect_configuration_id"]; got != want {
			return fmt.Errorf("End User Messaging SMS Account Default Protect Configuration: got %s, want %s", got, want)
		}

		return nil
	}
}

func testAccAccountDefaultProtectConfigurationConfig_basic(index int) string {
	return fmt.Sprintf(`
resource "aws_pinpointsmsvoicev2_account_default_protect_configuration" "test" {
  protect_configuration_id = aws_pinpointsmsvoicev2_protect_configuration.test[%[1]d].id
}

resource "aws_pinpointsmsvoicev2_protect_configuration" "test" {
  count = 2
}
`, index)
}
