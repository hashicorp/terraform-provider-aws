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

func testAccCollector_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v continuum.GetCollectorOutput
	resourceName := "aws_continuum_collector.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ContinuumServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCollectorDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccCollectorConfig_basic(),
				Check:  testAccCheckCollectorExists(ctx, t, resourceName, &v),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrType), knownvalue.StringExact("AWS_COLLECTOR")),
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

func testAccCollector_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var v continuum.GetCollectorOutput
	resourceName := "aws_continuum_collector.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ContinuumServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCollectorDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccCollectorConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCollectorExists(ctx, t, resourceName, &v),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfcontinuum.ResourceCollector, resourceName),
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

func testAccCheckCollectorExists(ctx context.Context, t *testing.T, n string, v *continuum.GetCollectorOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).ContinuumClient(ctx)

		output, err := tfcontinuum.FindCollectorByID(ctx, conn, rs.Primary.ID)
		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckCollectorDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).ContinuumClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_continuum_collector" {
				continue
			}

			_, err := tfcontinuum.FindCollectorByID(ctx, conn, rs.Primary.ID)
			if retry.NotFound(err) {
				continue
			}
			if err != nil {
				return err
			}

			return fmt.Errorf("Continuum Collector %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCollectorConfig_basic() string {
	return `
resource "aws_continuum_collector" "test" {}
`
}
