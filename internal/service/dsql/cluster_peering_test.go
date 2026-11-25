// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package dsql_test

import (
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccDSQLClusterPeering_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}
	resourceName1 := "aws_dsql_cluster_peering.test1"
	resourceName2 := "aws_dsql_cluster_peering.test2"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
			acctest.PreCheckMultipleRegion(t, 2)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.DSQLServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesMultipleRegions(ctx, t, 2),
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterPeeringConfig_basic(acctest.ThirdRegion(), false),
				Check: resource.ComposeAggregateTestCheckFunc(
					acctest.MatchResourceAttrRegionalARNRegion(ctx, resourceName1, "clusters.0", "dsql", acctest.AlternateRegion(), regexache.MustCompile(`cluster/.+$`)),
					acctest.MatchResourceAttrRegionalARNRegion(ctx, resourceName2, "clusters.0", "dsql", acctest.Region(), regexache.MustCompile(`cluster/.+$`)),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName1, plancheck.ResourceActionCreate),
						plancheck.ExpectResourceAction(resourceName2, plancheck.ResourceActionCreate),
					},
				},
			},
			{
				ResourceName:                         resourceName1,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName1, names.AttrIdentifier),
				ImportStateVerifyIdentifierAttribute: names.AttrIdentifier,
				ImportState:                          true,
				ImportStateVerify:                    true,
			},
		},
	})
}

func TestAccDSQLClusterPeering_deletionProtection(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}
	resourceName1 := "aws_dsql_cluster.test1"
	resourceName2 := "aws_dsql_cluster.test2"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
			acctest.PreCheckMultipleRegion(t, 2)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.DSQLServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesMultipleRegions(ctx, t, 2),
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterPeeringConfig_basic(acctest.ThirdRegion(), false),
			},
			{
				Config: testAccClusterPeeringConfig_basic(acctest.ThirdRegion(), true),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName1, plancheck.ResourceActionUpdate),
						plancheck.ExpectResourceAction(resourceName2, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName1, tfjsonpath.New("deletion_protection_enabled"), knownvalue.Bool(true)),
					statecheck.ExpectKnownValue(resourceName2, tfjsonpath.New("deletion_protection_enabled"), knownvalue.Bool(true)),
				},
			},
			{
				Config: testAccClusterPeeringConfig_basic(acctest.ThirdRegion(), false),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName1, plancheck.ResourceActionUpdate),
						plancheck.ExpectResourceAction(resourceName2, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName1, tfjsonpath.New("deletion_protection_enabled"), knownvalue.Bool(false)),
					statecheck.ExpectKnownValue(resourceName2, tfjsonpath.New("deletion_protection_enabled"), knownvalue.Bool(false)),
				},
			},
		},
	})
}

func testAccClusterPeeringConfig_basic(witnessRegion string, deletionProtection bool) string {
	return acctest.ConfigCompose(acctest.ConfigMultipleRegionProvider(2), fmt.Sprintf(`
resource "aws_dsql_cluster" "test1" {
  deletion_protection_enabled = %[2]t
  multi_region_properties {
    witness_region = %[1]q
  }
}

resource "aws_dsql_cluster_peering" "test1" {
  identifier     = aws_dsql_cluster.test1.identifier
  clusters       = [aws_dsql_cluster.test2.arn]
  witness_region = %[1]q
}

resource "aws_dsql_cluster" "test2" {
  provider = "awsalternate"

  deletion_protection_enabled = %[2]t
  multi_region_properties {
    witness_region = %[1]q
  }
}

resource "aws_dsql_cluster_peering" "test2" {
  provider = "awsalternate"

  identifier     = aws_dsql_cluster.test2.identifier
  clusters       = [aws_dsql_cluster.test1.arn]
  witness_region = %[1]q
}
`, witnessRegion, deletionProtection))
}
