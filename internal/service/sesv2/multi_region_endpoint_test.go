// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package sesv2_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/sesv2"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/names"

	tfsesv2 "github.com/hashicorp/terraform-provider-aws/internal/service/sesv2"
)

func TestAccSESV2MultiRegionEndpoint_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var multiregionendpoint sesv2.GetMultiRegionEndpointOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_sesv2_multi_region_endpoint.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SESV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckMultiRegionEndpointDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccMultiRegionEndpointConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckMultiRegionEndpointExists(ctx, t, resourceName, &multiregionendpoint),
					resource.TestCheckResourceAttr(resourceName, "endpoint_name", rName),
					resource.TestCheckResourceAttrSet(resourceName, "endpoint_id"),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, "READY"),
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

func TestAccSESV2MultiRegionEndpoint_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var multiregionendpoint sesv2.GetMultiRegionEndpointOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_sesv2_multi_region_endpoint.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SESV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckMultiRegionEndpointDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccMultiRegionEndpointConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckMultiRegionEndpointExists(ctx, t, resourceName, &multiregionendpoint),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfsesv2.ResourceMultiRegionEndpoint, resourceName),
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

func testAccCheckMultiRegionEndpointDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).SESV2Client(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_sesv2_multi_region_endpoint" {
				continue
			}

			_, err := tfsesv2.FindMultiRegionEndpointByName(ctx, conn, rs.Primary.ID)
			if retry.NotFound(err) {
				return nil
			}
			if err != nil {
				return create.Error(names.SESV2, create.ErrActionCheckingDestroyed, tfsesv2.ResNameMultiRegionEndpoint, rs.Primary.ID, err)
			}

			return create.Error(names.SESV2, create.ErrActionCheckingDestroyed, tfsesv2.ResNameMultiRegionEndpoint, rs.Primary.ID, errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccCheckMultiRegionEndpointExists(ctx context.Context, t *testing.T, name string, multiregionendpoint *sesv2.GetMultiRegionEndpointOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.SESV2, create.ErrActionCheckingExistence, tfsesv2.ResNameMultiRegionEndpoint, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.SESV2, create.ErrActionCheckingExistence, tfsesv2.ResNameMultiRegionEndpoint, name, errors.New("not set"))
		}

		conn := acctest.ProviderMeta(ctx, t).SESV2Client(ctx)

		resp, err := tfsesv2.FindMultiRegionEndpointByName(ctx, conn, rs.Primary.ID)
		if err != nil {
			return create.Error(names.SESV2, create.ErrActionCheckingExistence, tfsesv2.ResNameMultiRegionEndpoint, rs.Primary.ID, err)
		}

		*multiregionendpoint = *resp

		return nil
	}
}

func testAccPreCheck(ctx context.Context, t *testing.T) {
	conn := acctest.ProviderMeta(ctx, t).SESV2Client(ctx)

	input := &sesv2.ListMultiRegionEndpointsInput{}

	_, err := conn.ListMultiRegionEndpoints(ctx, input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}
	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccMultiRegionEndpointConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_sesv2_multi_region_endpoint" "test" {
  endpoint_name = %[1]q

  details {
    routes_details {
      region = data.aws_region.secondary.name
    }
  }
}

data "aws_region" "secondary" {}
`, rName)
}
