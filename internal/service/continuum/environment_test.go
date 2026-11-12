// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package continuum_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/continuum"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfcontinuum "github.com/hashicorp/terraform-provider-aws/internal/service/continuum"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccEnvironment_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v continuum.GetEnvironmentOutput
	resourceName := "aws_continuum_environment.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ContinuumServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEnvironmentDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccEnvironmentConfig_basic(rName, rName, "PT0H"),
				Check:  testAccCheckEnvironmentExists(ctx, t, resourceName, &v),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrName), knownvalue.StringExact(rName)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("scan_frequency"), knownvalue.StringExact("PT0H")),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrStatus), knownvalue.StringExact("ACTIVE")),
				},
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccEnvironment_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var v continuum.GetEnvironmentOutput
	resourceName := "aws_continuum_environment.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ContinuumServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEnvironmentDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccEnvironmentConfig_basic(rName, rName, "PT0H"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEnvironmentExists(ctx, t, resourceName, &v),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfcontinuum.ResourceEnvironment, resourceName),
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

func testAccEnvironment_update(t *testing.T) {
	ctx := acctest.Context(t)
	var v continuum.GetEnvironmentOutput
	resourceName := "aws_continuum_environment.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	rNameUpdated := rName + "-updated"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ContinuumServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEnvironmentDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccEnvironmentConfig_basic(rName, rName, "PT0H"),
				Check:  testAccCheckEnvironmentExists(ctx, t, resourceName, &v),
			},
			{
				Config: testAccEnvironmentConfig_basic(rName, rNameUpdated, "P7D"),
				Check:  testAccCheckEnvironmentExists(ctx, t, resourceName, &v),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrName), knownvalue.StringExact(rNameUpdated)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("scan_frequency"), knownvalue.StringExact("P7D")),
				},
			},
		},
	})
}

func testAccCheckEnvironmentExists(ctx context.Context, t *testing.T, n string, v *continuum.GetEnvironmentOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).ContinuumClient(ctx)

		output, err := tfcontinuum.FindEnvironmentByID(ctx, conn, rs.Primary.ID)
		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckEnvironmentDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).ContinuumClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_continuum_environment" {
				continue
			}

			_, err := tfcontinuum.FindEnvironmentByID(ctx, conn, rs.Primary.ID)
			if retry.NotFound(err) {
				continue
			}
			if err != nil {
				return err
			}

			return fmt.Errorf("Continuum Environment %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccEnvironmentConfig_basic(rName, envName, scanFrequency string) string {
	return acctest.ConfigCompose(testAccIntegrationConfig_basic(rName), fmt.Sprintf(`
resource "aws_continuum_environment" "test" {
  name           = %[1]q
  scan_frequency = %[2]q

  s3_source {
    integration_id = aws_continuum_integration.test.id
    code_prefix    = "code/"
  }
}
`, envName, scanFrequency))
}
