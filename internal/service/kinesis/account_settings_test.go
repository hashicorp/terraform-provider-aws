// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package kinesis_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfkinesis "github.com/hashicorp/terraform-provider-aws/internal/service/kinesis"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccKinesisAccountSettings_serial(t *testing.T) {
	acctest.Skip(t, "Kinesis Data Streams On-demand Advantage has a minimum 24 hour commitment and costs > $100/day")
	t.Parallel()

	testCases := map[string]func(t *testing.T){
		acctest.CtBasic: testAccAccountSettings_basic,
		"enabled":       testAccAccountSettings_enabled,
		"Identity":      testAccKinesisAccountSettings_identitySerial,
	}

	acctest.RunSerialTests1Level(t, testCases, 0)
}

func testAccAccountSettings_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_kinesis_account_settings.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.KinesisServiceID),
		CheckDestroy:             acctest.CheckDestroyNoop,
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.StaticDirectory("testdata/AccountSettings/basic/"),
				ConfigVariables: config.Variables{},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAccountSettingsExists(ctx, t, resourceName),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
		},
	})
}

func testAccAccountSettings_enabled(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_kinesis_account_settings.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.KinesisServiceID),
		CheckDestroy:             acctest.CheckDestroyNoop,
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.StaticDirectory("testdata/AccountSettings/status/"),
				ConfigVariables: config.Variables{
					"status": config.StringVariable("ENABLED"),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAccountSettingsExists(ctx, t, resourceName),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
			{
				ConfigDirectory: config.StaticDirectory("testdata/AccountSettings/status/"),
				ConfigVariables: config.Variables{
					"status": config.StringVariable("ENABLED"),
				},
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				ConfigDirectory: config.StaticDirectory("testdata/AccountSettings/status/"),
				ConfigVariables: config.Variables{
					"status": config.StringVariable("DISABLED"),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAccountSettingsExists(ctx, t, resourceName),
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

func testAccCheckAccountSettingsExists(ctx context.Context, t *testing.T, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).KinesisClient(ctx)

		_, err := tfkinesis.FindAccountSettings(ctx, conn)

		return err
	}
}
