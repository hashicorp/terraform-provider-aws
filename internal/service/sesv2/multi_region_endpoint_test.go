// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package sesv2_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfsesv2 "github.com/hashicorp/terraform-provider-aws/internal/service/sesv2"
	"github.com/hashicorp/terraform-provider-aws/names"
)

const (
	ResNameMultiRegionEndpoint = "Multi Region Endpoint"
)

func TestAccSESV2MultiRegionEndpoint_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_sesv2_multi_region_endpoint.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SESV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckMultiRegionEndpointDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccMultiRegionEndpointConfig_basic(rName, acctest.AlternateRegion()),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckMultiRegionEndpointExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "endpoint_name", rName),
					resource.TestCheckResourceAttrSet(resourceName, "endpoint_id"),
					resource.TestCheckResourceAttr(resourceName, "details.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "details.0.routes_details.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "details.0.routes_details.0.region", acctest.AlternateRegion()),
				),
			},
		},
	})
}

func TestAccSESV2MultiRegionEndpoint_tags(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_sesv2_multi_region_endpoint.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SESV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckMultiRegionEndpointDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccMultiRegionEndpointConfig_tags(rName, acctest.AlternateRegion(), acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckMultiRegionEndpointExists(ctx, t, resourceName),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.MapExact(map[string]knownvalue.Check{
						acctest.CtKey1: knownvalue.StringExact(acctest.CtValue1),
					})),
				},
			},
		},
	})
}

func TestAccSESV2MultiRegionEndpoint_update(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_sesv2_multi_region_endpoint.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckMultipleRegion(t, 3)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SESV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckMultiRegionEndpointDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccMultiRegionEndpointConfig_basic(rName, acctest.AlternateRegion()),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckMultiRegionEndpointExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "endpoint_name", rName),
				),
			},
			{
				Config: testAccMultiRegionEndpointConfig_basic(rName, acctest.ThirdRegion()),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionDestroyBeforeCreate),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckMultiRegionEndpointExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "endpoint_name", rName),
				),
			},
		},
	})
}

func TestAccSESV2MultiRegionEndpoint_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_sesv2_multi_region_endpoint.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SESV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckMultiRegionEndpointDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccMultiRegionEndpointConfig_basic(rName, acctest.AlternateRegion()),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckMultiRegionEndpointExists(ctx, t, resourceName),
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

			_, err := tfsesv2.FindMultiRegionEndpointByName(ctx, conn, rs.Primary.Attributes["endpoint_name"])
			if retry.NotFound(err) {
				return nil
			}
			if err != nil {
				return create.Error(names.SESV2, create.ErrActionCheckingDestroyed, ResNameMultiRegionEndpoint, rs.Primary.Attributes["endpoint_name"], err)
			}

			return create.Error(names.SESV2, create.ErrActionCheckingDestroyed, ResNameMultiRegionEndpoint, rs.Primary.Attributes["endpoint_name"], errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccCheckMultiRegionEndpointExists(ctx context.Context, t *testing.T, name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.SESV2, create.ErrActionCheckingExistence, ResNameMultiRegionEndpoint, name, errors.New("not found"))
		}

		conn := acctest.ProviderMeta(ctx, t).SESV2Client(ctx)

		_, err := tfsesv2.FindMultiRegionEndpointByName(ctx, conn, rs.Primary.Attributes["endpoint_name"])

		return err
	}
}

func testAccMultiRegionEndpointConfig_basic(rName, alternateRegion string) string {
	return fmt.Sprintf(`
resource "aws_sesv2_multi_region_endpoint" "test" {
  endpoint_name = %[1]q

  details {
    routes_details {
      region = %[2]q
    }
  }
}
`, rName, alternateRegion)
}

func testAccMultiRegionEndpointConfig_tags(rName, alternateRegion, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_sesv2_multi_region_endpoint" "test" {
  endpoint_name = %[1]q

  details {
    routes_details {
      region = %[2]q
    }
  }

  tags = {
    %[3]q = %[4]q
  }
}
`, rName, alternateRegion, tagKey1, tagValue1)
}
