// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package resiliencehubv2_test

import (
	"context"
	"fmt"
	"testing"

	awstypes "github.com/aws/aws-sdk-go-v2/service/resiliencehubv2/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfresiliencehubv2 "github.com/hashicorp/terraform-provider-aws/internal/service/resiliencehubv2"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccResilienceHubV2System_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var system awstypes.System
	rName := acctest.RandomWithPrefix(t, "tf-acc-test")
	resourceName := "aws_resiliencehubv2_system.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ResilienceHubV2),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSystemDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccSystemConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSystemExists(ctx, t, resourceName, &system),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrARN),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: names.AttrARN,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, names.AttrARN),
			},
		},
	})
}

func TestAccResilienceHubV2System_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var system awstypes.System
	rName := acctest.RandomWithPrefix(t, "tf-acc-test")
	resourceName := "aws_resiliencehubv2_system.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ResilienceHubV2),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSystemDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccSystemConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSystemExists(ctx, t, resourceName, &system),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfresiliencehubv2.ResourceSystem, resourceName),
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

func TestAccResilienceHubV2System_update(t *testing.T) {
	ctx := acctest.Context(t)
	var system awstypes.System
	rName := acctest.RandomWithPrefix(t, "tf-acc-test")
	resourceName := "aws_resiliencehubv2_system.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ResilienceHubV2),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSystemDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccSystemConfig_full(rName, "initial description"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSystemExists(ctx, t, resourceName, &system),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "initial description"),
				),
			},
			{
				Config: testAccSystemConfig_full(rName, "updated description"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSystemExists(ctx, t, resourceName, &system),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "updated description"),
				),
			},
		},
	})
}

func testAccCheckSystemDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).ResilienceHubV2Client(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_resiliencehubv2_system" {
				continue
			}

			_, err := tfresiliencehubv2.FindSystemByARN(ctx, conn, rs.Primary.Attributes[names.AttrARN])
			if err == nil {
				return fmt.Errorf("Resilience Hub V2 System %s still exists", rs.Primary.Attributes[names.AttrARN])
			}
		}

		return nil
	}
}

func testAccCheckSystemExists(ctx context.Context, t *testing.T, n string, v *awstypes.System) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("System not found: %s", n)
		}

		if rs.Primary.Attributes[names.AttrARN] == "" {
			return fmt.Errorf("No Resilience Hub V2 System ARN is set")
		}

		conn := acctest.ProviderMeta(ctx, t).ResilienceHubV2Client(ctx)

		output, err := tfresiliencehubv2.FindSystemByARN(ctx, conn, rs.Primary.Attributes[names.AttrARN])
		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccSystemConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_resiliencehubv2_system" "test" {
  name = %[1]q
}
`, rName)
}

func testAccSystemConfig_full(rName, description string) string {
	return fmt.Sprintf(`
resource "aws_resiliencehubv2_system" "test" {
  name        = %[1]q
  description = %[2]q
}
`, rName, description)
}
