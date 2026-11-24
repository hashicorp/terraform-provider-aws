// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package ec2_test

import (
	"context"
	"fmt"
	"testing"

	awstypes "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfstatecheck "github.com/hashicorp/terraform-provider-aws/internal/acctest/statecheck"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccEC2PublicIPv4Pool_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.PublicIpv4Pool
	resourceName := "aws_ec2_public_ipv4_pool.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPublicIPv4PoolDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccPublicIPv4PoolConfig_basic(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckPublicIPv4PoolExists(ctx, t, resourceName, &v),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
						plancheck.ExpectUnknownValue(resourceName, tfjsonpath.New(names.AttrDescription)),
						plancheck.ExpectUnknownValue(resourceName, tfjsonpath.New("pool_address_ranges")),
						plancheck.ExpectUnknownValue(resourceName, tfjsonpath.New("total_address_count")),
						plancheck.ExpectUnknownValue(resourceName, tfjsonpath.New("total_available_address_count")),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					tfstatecheck.ExpectRegionalARNFormat(resourceName, tfjsonpath.New(names.AttrARN), "ec2", "ipv4pool-ec2/{pool_id}"),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("pool_id"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrDescription), knownvalue.StringExact("")),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("network_border_group"), knownvalue.StringExact(acctest.Region())),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("pool_address_ranges"), knownvalue.ListExact([]knownvalue.Check{})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.Null()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTagsAll), knownvalue.MapExact(map[string]knownvalue.Check{})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("total_address_count"), knownvalue.Int64Exact(0)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("total_available_address_count"), knownvalue.Int64Exact(0)),
				},
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "pool_id"),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "pool_id",
			},
			{
				RefreshState: true,
				RefreshPlanChecks: resource.RefreshPlanChecks{
					PostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
			},
		},
	})
}

func TestAccEC2PublicIPv4Pool_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.PublicIpv4Pool
	resourceName := "aws_ec2_public_ipv4_pool.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPublicIPv4PoolDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccPublicIPv4PoolConfig_basic(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckPublicIPv4PoolExists(ctx, t, resourceName, &v),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfec2.ResourcePublicIPv4Pool, resourceName),
				),
				ExpectNonEmptyPlan: true,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
		},
	})
}

func TestAccEC2PublicIPv4Pool_networkBorderGroup(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.PublicIpv4Pool
	resourceName := "aws_ec2_public_ipv4_pool.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPublicIPv4PoolDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccPublicIPv4PoolConfig_networkBorderGroup(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckPublicIPv4PoolExists(ctx, t, resourceName, &v),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("network_border_group"), knownvalue.StringExact(acctest.Region())),
				},
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "pool_id"),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "pool_id",
			},
			{
				RefreshState: true,
				RefreshPlanChecks: resource.RefreshPlanChecks{
					PostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
			},
		},
	})
}

func testAccCheckPublicIPv4PoolDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).EC2Client(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_ec2_public_ipv4_pool" {
				continue
			}

			id := rs.Primary.Attributes["pool_id"]
			_, err := tfec2.FindPublicIPv4PoolByID(ctx, conn, id)
			if retry.NotFound(err) {
				continue
			}
			if err != nil {
				return err
			}

			return fmt.Errorf("Public IPv4 Pool %s still exists", id)
		}

		return nil
	}
}

func testAccCheckPublicIPv4PoolExists(ctx context.Context, t *testing.T, n string, v *awstypes.PublicIpv4Pool) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).EC2Client(ctx)

		id := rs.Primary.Attributes["pool_id"]
		output, err := tfec2.FindPublicIPv4PoolByID(ctx, conn, id)
		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccPublicIPv4PoolConfig_basic() string {
	return `
resource "aws_ec2_public_ipv4_pool" "test" {}
`
}

func testAccPublicIPv4PoolConfig_networkBorderGroup() string {
	return `
resource "aws_ec2_public_ipv4_pool" "test" {
  network_border_group = data.aws_region.current.region
}

data "aws_region" "current" {}
`
}
