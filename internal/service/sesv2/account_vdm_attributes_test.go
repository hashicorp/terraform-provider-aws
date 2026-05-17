// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package sesv2_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/sesv2/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfsesv2 "github.com/hashicorp/terraform-provider-aws/internal/service/sesv2"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccSESV2AccountVDMAttributes_serial(t *testing.T) {
	t.Parallel()

	testCases := map[string]func(t *testing.T){
		acctest.CtBasic:           testAccAccountVDMAttributes_basic,
		acctest.CtDisappears:      testAccAccountVDMAttributes_disappears,
		"engagementMetrics":       testAccAccountVDMAttributes_engagementMetrics,
		"optimizedSharedDelivery": testAccAccountVDMAttributes_optimizedSharedDelivery,
	}

	acctest.RunSerialTests1Level(t, testCases, 0)
}

func testAccAccountVDMAttributes_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_sesv2_account_vdm_attributes.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SESV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAccountVDMAttributesDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccAccountVDMAttributesConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "vdm_enabled", string(types.FeatureStatusEnabled)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccAccountVDMAttributes_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_sesv2_account_vdm_attributes.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SESV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAccountVDMAttributesDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccAccountVDMAttributesConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckSDKResourceDisappears(ctx, t, tfsesv2.ResourceAccountVDMAttributes(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccAccountVDMAttributes_engagementMetrics(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_sesv2_account_vdm_attributes.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SESV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAccountVDMAttributesDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccAccountVDMAttributesConfig_engagementMetrics(string(types.FeatureStatusEnabled)),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "dashboard_attributes.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "dashboard_attributes.0.engagement_metrics", string(types.FeatureStatusEnabled)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAccountVDMAttributesConfig_engagementMetrics(string(types.FeatureStatusDisabled)),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "dashboard_attributes.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "dashboard_attributes.0.engagement_metrics", string(types.FeatureStatusDisabled)),
				),
			},
		},
	})
}

func testAccAccountVDMAttributes_optimizedSharedDelivery(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_sesv2_account_vdm_attributes.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SESV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAccountVDMAttributesDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccAccountVDMAttributesConfig_optimizedSharedDelivery(string(types.FeatureStatusEnabled)),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "guardian_attributes.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "guardian_attributes.0.optimized_shared_delivery", string(types.FeatureStatusEnabled)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAccountVDMAttributesConfig_optimizedSharedDelivery(string(types.FeatureStatusDisabled)),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "guardian_attributes.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "guardian_attributes.0.optimized_shared_delivery", string(types.FeatureStatusDisabled)),
				),
			},
		},
	})
}

func testAccCheckAccountVDMAttributesDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).SESV2Client(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_sesv2_account_vdm_attributes" {
				continue
			}

			output, err := tfsesv2.FindAccountVDMAttributes(ctx, conn)

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			if output.VdmEnabled == types.FeatureStatusDisabled {
				continue
			}

			return fmt.Errorf("SESv2 Account VDM Attributes %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccAccountVDMAttributesConfig_basic() string {
	return `
resource "aws_sesv2_account_vdm_attributes" "test" {
  vdm_enabled = "ENABLED"
}
`
}

func testAccAccountVDMAttributesConfig_engagementMetrics(engagementMetrics string) string {
	return fmt.Sprintf(`
resource "aws_sesv2_account_vdm_attributes" "test" {
  vdm_enabled = "ENABLED"

  dashboard_attributes {
    engagement_metrics = %[1]q
  }
}
`, engagementMetrics)
}

func testAccAccountVDMAttributesConfig_optimizedSharedDelivery(optimizedSharedDelivery string) string {
	return fmt.Sprintf(`
resource "aws_sesv2_account_vdm_attributes" "test" {
  vdm_enabled = "ENABLED"

  guardian_attributes {
    optimized_shared_delivery = %[1]q
  }
}
`, optimizedSharedDelivery)
}
