// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package ec2_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccEC2CapacityManagerSettings_serial(t *testing.T) {
	t.Parallel()

	testCases := map[string]func(*testing.T){
		acctest.CtBasic:        testAccEC2CapacityManagerSettings_basic,
		"update":               testAccEC2CapacityManagerSettings_update,
		"organizationsAccess":  testAccEC2CapacityManagerSettings_organizationsAccess,
		"invalidConfiguration": testAccEC2CapacityManagerSettings_invalidConfiguration,
		"Identity":             testAccEC2CapacityManagerSettings_identitySerial,
	}

	acctest.RunSerialTests1Level(t, testCases, 0)
}

func testAccEC2CapacityManagerSettings_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_ec2_capacity_manager_settings.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCapacityManagerSettingsDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccCapacityManagerSettingsConfig_basic(true),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCapacityManagerSettingsExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrEnabled, acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "organizations_access", acctest.CtFalse),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrRegion),
					resource.TestCheckResourceAttr(resourceName, names.AttrID, acctest.Region()),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateId:                        acctest.Region(),
				ImportStateVerifyIdentifierAttribute: names.AttrRegion,
				ImportStateVerify:                    true,
			},
		},
	})
}

func testAccEC2CapacityManagerSettings_update(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_ec2_capacity_manager_settings.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCapacityManagerSettingsDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccCapacityManagerSettingsConfig_basic(false),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCapacityManagerSettingsDisabled(ctx, t),
					resource.TestCheckResourceAttr(resourceName, names.AttrEnabled, acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "organizations_access", acctest.CtFalse),
				),
			},
			{
				Config: testAccCapacityManagerSettingsConfig_basic(true),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCapacityManagerSettingsExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrEnabled, acctest.CtTrue),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
			},
			{
				Config: testAccCapacityManagerSettingsConfig_basic(false),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCapacityManagerSettingsDisabled(ctx, t),
					resource.TestCheckResourceAttr(resourceName, names.AttrEnabled, acctest.CtFalse),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
			},
		},
	})
}

func testAccEC2CapacityManagerSettings_organizationsAccess(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_ec2_capacity_manager_settings.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckOrganizationManagementAccount(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCapacityManagerSettingsDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccCapacityManagerSettingsConfig_organizationsAccess(true, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCapacityManagerSettingsExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrEnabled, acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "organizations_access", acctest.CtTrue),
				),
			},
			{
				Config: testAccCapacityManagerSettingsConfig_organizationsAccess(true, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCapacityManagerSettingsExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "organizations_access", acctest.CtFalse),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
			},
		},
	})
}

func testAccEC2CapacityManagerSettings_invalidConfiguration(t *testing.T) {
	ctx := acctest.Context(t)

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccCapacityManagerSettingsConfig_organizationsAccess(false, true),
				ExpectError: regexache.MustCompile(`organizations_access cannot be true when enabled is false`),
			},
		},
	})
}

func testAccCheckCapacityManagerSettingsExists(ctx context.Context, t *testing.T, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).EC2Client(ctx)

		output, err := tfec2.FindCapacityManagerAttributes(ctx, conn)
		if err != nil {
			return err
		}

		if output.CapacityManagerStatus != awstypes.CapacityManagerStatusEnabled {
			return errors.New("EC2 Capacity Manager not enabled")
		}

		return nil
	}
}

func testAccCheckCapacityManagerSettingsDisabled(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).EC2Client(ctx)

		output, err := tfec2.FindCapacityManagerAttributes(ctx, conn)

		if retry.NotFound(err) {
			return nil
		}

		if err != nil {
			return err
		}

		if output.CapacityManagerStatus != awstypes.CapacityManagerStatusDisabled {
			return errors.New("EC2 Capacity Manager still enabled")
		}

		return nil
	}
}

func testAccCheckCapacityManagerSettingsDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return testAccCheckCapacityManagerSettingsDisabled(ctx, t)
}

func testAccCapacityManagerSettingsConfig_basic(enabled bool) string {
	return fmt.Sprintf(`
resource "aws_ec2_capacity_manager_settings" "test" {
  enabled = %[1]t
}
`, enabled)
}

func testAccCapacityManagerSettingsConfig_organizationsAccess(enabled, organizationsAccess bool) string {
	return fmt.Sprintf(`
resource "aws_ec2_capacity_manager_settings" "test" {
  enabled              = %[1]t
  organizations_access = %[2]t
}
`, enabled, organizationsAccess)
}
