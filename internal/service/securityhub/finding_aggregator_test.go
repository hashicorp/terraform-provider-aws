// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package securityhub_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfsecurityhub "github.com/hashicorp/terraform-provider-aws/internal/service/securityhub"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccFindingAggregator_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_securityhub_finding_aggregator.test_aggregator"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SecurityHubServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFindingAggregatorDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccFindingAggregatorConfig_allRegions(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFindingAggregatorExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "linking_mode", "ALL_REGIONS"),
					resource.TestCheckNoResourceAttr(resourceName, "specified_regions"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccFindingAggregatorConfig_specifiedRegions(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFindingAggregatorExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "linking_mode", "SPECIFIED_REGIONS"),
					resource.TestCheckResourceAttr(resourceName, "specified_regions.#", acctest.Ct3),
				),
			},
			{
				Config: testAccFindingAggregatorConfig_allRegionsExceptSpecified(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFindingAggregatorExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "linking_mode", "ALL_REGIONS_EXCEPT_SPECIFIED"),
					resource.TestCheckResourceAttr(resourceName, "specified_regions.#", acctest.Ct2),
				),
			},
		},
	})
}

func testAccFindingAggregator_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_securityhub_finding_aggregator.test_aggregator"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SecurityHubServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFindingAggregatorDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccFindingAggregatorConfig_allRegions(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFindingAggregatorExists(ctx, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfsecurityhub.ResourceFindingAggregator(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckFindingAggregatorExists(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).SecurityHubClient(ctx)

		_, err := tfsecurityhub.FindFindingAggregatorByARN(ctx, conn, rs.Primary.ID)

		return err
	}
}

func testAccCheckFindingAggregatorDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).SecurityHubClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_securityhub_finding_aggregator" {
				continue
			}

			_, err := tfsecurityhub.FindFindingAggregatorByARN(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Security Hub Finding Aggregator (%s) still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccFindingAggregatorConfig_allRegions() string {
	return `
resource "aws_securityhub_account" "example" {}

resource "aws_securityhub_finding_aggregator" "test_aggregator" {
  linking_mode = "ALL_REGIONS"

  depends_on = [aws_securityhub_account.example]
}
`
}

func testAccFindingAggregatorConfig_specifiedRegions() string {
	return fmt.Sprintf(`
resource "aws_securityhub_account" "example" {}

resource "aws_securityhub_finding_aggregator" "test_aggregator" {
  linking_mode      = "SPECIFIED_REGIONS"
  specified_regions = ["%s", "%s", "%s"]

  depends_on = [aws_securityhub_account.example]
}
`, endpoints.EuWest1RegionID, endpoints.EuWest2RegionID, endpoints.UsEast1RegionID)
}

func testAccFindingAggregatorConfig_allRegionsExceptSpecified() string {
	return fmt.Sprintf(`
resource "aws_securityhub_account" "example" {}

resource "aws_securityhub_finding_aggregator" "test_aggregator" {
  linking_mode      = "ALL_REGIONS_EXCEPT_SPECIFIED"
  specified_regions = ["%s", "%s"]

  depends_on = [aws_securityhub_account.example]
}
`, endpoints.EuWest1RegionID, endpoints.EuWest2RegionID)
}
