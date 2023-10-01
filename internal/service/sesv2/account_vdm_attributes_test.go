// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package sesv2_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/sesv2/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/names"

	tfsesv2 "github.com/hashicorp/terraform-provider-aws/internal/service/sesv2"
)

func TestAccSESV2AccountVdmAttributes_serial(t *testing.T) {
	t.Parallel()

	testCases := map[string]func(t *testing.T){
		"basic":                   testAccSESV2AccountVdmAttributes_basic,
		"disappears":              testAccSESV2AccountVdmAttributes_disappears,
		"engagementMetrics":       testAccSESV2AccountVdmAttributes_engagementMetrics,
		"optimizedSharedDelivery": testAccSESV2AccountVdmAttributes_optimizedSharedDelivery,
	}

	acctest.RunSerialTests1Level(t, testCases, 0)
}

func testAccSESV2AccountVdmAttributes_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_sesv2_account_vdm_attributes.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SESV2EndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAccountVdmAttributesDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAccountVdmAttributesConfig_basic(),
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

func testAccSESV2AccountVdmAttributes_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_sesv2_account_vdm_attributes.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SESV2EndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAccountVdmAttributesDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAccountVdmAttributesConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfsesv2.ResourceAccountVDMAttributes(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccSESV2AccountVdmAttributes_engagementMetrics(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_sesv2_account_vdm_attributes.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SESV2EndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAccountVdmAttributesDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAccountVdmAttributesConfig_engagementMetrics(string(types.FeatureStatusEnabled)),
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
				Config: testAccAccountVdmAttributesConfig_engagementMetrics(string(types.FeatureStatusDisabled)),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "dashboard_attributes.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "dashboard_attributes.0.engagement_metrics", string(types.FeatureStatusDisabled)),
				),
			},
		},
	})
}

func testAccSESV2AccountVdmAttributes_optimizedSharedDelivery(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_sesv2_account_vdm_attributes.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SESV2EndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAccountVdmAttributesDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAccountVdmAttributesConfig_optimizedSharedDelivery(string(types.FeatureStatusEnabled)),
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
				Config: testAccAccountVdmAttributesConfig_optimizedSharedDelivery(string(types.FeatureStatusDisabled)),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "guardian_attributes.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "guardian_attributes.0.optimized_shared_delivery", string(types.FeatureStatusDisabled)),
				),
			},
		},
	})
}

func testAccCheckAccountVdmAttributesDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).SESV2Client(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_sesv2_account_vdm_attributes" {
				continue
			}

			out, err := tfsesv2.FindAccountVDMAttributes(ctx, conn)
			if err != nil {
				return err
			}

			if out.VdmEnabled == types.FeatureStatusDisabled {
				return nil
			}

			return create.Error(names.SESV2, create.ErrActionCheckingDestroyed, tfsesv2.ResNameAccountVDMAttributes, rs.Primary.ID, errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccAccountVdmAttributesConfig_basic() string {
	return `
resource "aws_sesv2_account_vdm_attributes" "test" {
	vdm_enabled = "ENABLED"
}
`
}

func testAccAccountVdmAttributesConfig_engagementMetrics(engagementMetrics string) string {
	return fmt.Sprintf(`
resource "aws_sesv2_account_vdm_attributes" "test" {
	vdm_enabled = "ENABLED"

	dashboard_attributes {
		engagement_metrics = %[1]q
	}
}
`, engagementMetrics)
}

func testAccAccountVdmAttributesConfig_optimizedSharedDelivery(optimizedSharedDelivery string) string {
	return fmt.Sprintf(`
resource "aws_sesv2_account_vdm_attributes" "test" {
	vdm_enabled = "ENABLED"

	guardian_attributes {
		optimized_shared_delivery = %[1]q
	}
}
`, optimizedSharedDelivery)
}
