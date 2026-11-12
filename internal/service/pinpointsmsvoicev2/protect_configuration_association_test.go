// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package pinpointsmsvoicev2_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	awstypes "github.com/aws/aws-sdk-go-v2/service/pinpointsmsvoicev2/types"
	"github.com/hashicorp/terraform-plugin-testing/compare"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfpinpointsmsvoicev2 "github.com/hashicorp/terraform-provider-aws/internal/service/pinpointsmsvoicev2"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccPinpointSMSVoiceV2ProtectConfigurationAssociation_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.ConfigurationSetInformation
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_pinpointsmsvoicev2_protect_configuration_association.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckProtectConfiguration(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.PinpointSMSVoiceV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckProtectConfigurationAssociationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccProtectConfigurationAssociationConfig_basic(rName, 0),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckProtectConfigurationAssociationExists(ctx, t, resourceName, &v),
					testAccCheckProtectConfigurationAssociationProtectConfigurationID(&v, "aws_pinpointsmsvoicev2_protect_configuration.test.0"),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.CompareValuePairs(resourceName, tfjsonpath.New("configuration_set_name"), "aws_pinpointsmsvoicev2_configuration_set.test", tfjsonpath.New(names.AttrName), compare.ValuesSame()),
					statecheck.CompareValuePairs(resourceName, tfjsonpath.New("protect_configuration_id"), "aws_pinpointsmsvoicev2_protect_configuration.test[0]", tfjsonpath.New(names.AttrID), compare.ValuesSame()),
				},
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "configuration_set_name"),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "configuration_set_name",
			},
		},
	})
}

func TestAccPinpointSMSVoiceV2ProtectConfigurationAssociation_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.ConfigurationSetInformation
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_pinpointsmsvoicev2_protect_configuration_association.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckProtectConfiguration(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.PinpointSMSVoiceV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckProtectConfigurationAssociationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccProtectConfigurationAssociationConfig_basic(rName, 0),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckProtectConfigurationAssociationExists(ctx, t, resourceName, &v),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfpinpointsmsvoicev2.ResourceProtectConfigurationAssociation, resourceName),
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

func TestAccPinpointSMSVoiceV2ProtectConfigurationAssociation_disappears_configurationSet(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.ConfigurationSetInformation
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_pinpointsmsvoicev2_protect_configuration_association.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckProtectConfiguration(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.PinpointSMSVoiceV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckProtectConfigurationAssociationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccProtectConfigurationAssociationConfig_basic(rName, 0),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckProtectConfigurationAssociationExists(ctx, t, resourceName, &v),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfpinpointsmsvoicev2.ResourceConfigurationSet, "aws_pinpointsmsvoicev2_configuration_set.test"),
				),
				ExpectNonEmptyPlan: true,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("aws_pinpointsmsvoicev2_configuration_set.test", plancheck.ResourceActionCreate),
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
		},
	})
}

func TestAccPinpointSMSVoiceV2ProtectConfigurationAssociation_update(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.ConfigurationSetInformation
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_pinpointsmsvoicev2_protect_configuration_association.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckProtectConfiguration(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.PinpointSMSVoiceV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckProtectConfigurationAssociationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccProtectConfigurationAssociationConfig_basic(rName, 0),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckProtectConfigurationAssociationExists(ctx, t, resourceName, &v),
					testAccCheckProtectConfigurationAssociationProtectConfigurationID(&v, "aws_pinpointsmsvoicev2_protect_configuration.test.0"),
				),
			},
			{
				Config: testAccProtectConfigurationAssociationConfig_basic(rName, 1),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckProtectConfigurationAssociationExists(ctx, t, resourceName, &v),
					testAccCheckProtectConfigurationAssociationProtectConfigurationID(&v, "aws_pinpointsmsvoicev2_protect_configuration.test.1"),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.CompareValuePairs(resourceName, tfjsonpath.New("protect_configuration_id"), "aws_pinpointsmsvoicev2_protect_configuration.test[1]", tfjsonpath.New(names.AttrID), compare.ValuesSame()),
				},
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
			},
		},
	})
}

func testAccCheckProtectConfigurationAssociationDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_pinpointsmsvoicev2_protect_configuration_association" {
				continue
			}

			ctx := acctest.NewTestResourceContext(ctx, rs.Type, rs.Primary.Attributes[names.AttrRegion])
			conn := acctest.ProviderMeta(ctx, t).PinpointSMSVoiceV2Client(ctx)
			name := rs.Primary.Attributes["configuration_set_name"]
			_, err := tfpinpointsmsvoicev2.FindProtectConfigurationAssociationByConfigurationSetName(ctx, conn, name)

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("End User Messaging SMS Protect Configuration Association %s still exists", name)
		}

		return nil
	}
}

func testAccCheckProtectConfigurationAssociationExists(ctx context.Context, t *testing.T, n string, v *awstypes.ConfigurationSetInformation) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		ctx := acctest.NewTestResourceContext(ctx, rs.Type, rs.Primary.Attributes[names.AttrRegion])
		conn := acctest.ProviderMeta(ctx, t).PinpointSMSVoiceV2Client(ctx)

		output, err := tfpinpointsmsvoicev2.FindProtectConfigurationAssociationByConfigurationSetName(ctx, conn, rs.Primary.Attributes["configuration_set_name"])
		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

// testAccCheckProtectConfigurationAssociationProtectConfigurationID checks the association stored by AWS, independent of state.
func testAccCheckProtectConfigurationAssociationProtectConfigurationID(v *awstypes.ConfigurationSetInformation, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if got, want := aws.ToString(v.ProtectConfigurationId), rs.Primary.Attributes[names.AttrID]; got != want {
			return fmt.Errorf("End User Messaging SMS Configuration Set %s protect configuration: got %s, want %s", aws.ToString(v.ConfigurationSetName), got, want)
		}

		return nil
	}
}

func testAccProtectConfigurationAssociationConfig_basic(rName string, index int) string {
	return fmt.Sprintf(`
resource "aws_pinpointsmsvoicev2_protect_configuration_association" "test" {
  configuration_set_name   = aws_pinpointsmsvoicev2_configuration_set.test.name
  protect_configuration_id = aws_pinpointsmsvoicev2_protect_configuration.test[%[2]d].id
}

resource "aws_pinpointsmsvoicev2_configuration_set" "test" {
  name = %[1]q
}

resource "aws_pinpointsmsvoicev2_protect_configuration" "test" {
  count = 2
}
`, rName, index)
}
