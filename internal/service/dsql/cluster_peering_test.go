// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package dsql_test

import (
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
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
				Config: testAccClusterPeeringConfig_basic(acctest.ThirdRegion()),
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

func testAccClusterPeeringConfig_basic(witnessRegion string) string {
	return acctest.ConfigCompose(acctest.ConfigMultipleRegionProvider(2), fmt.Sprintf(`
resource "aws_dsql_cluster" "test1" {
  deletion_protection_enabled = false
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

  deletion_protection_enabled = false
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
`, witnessRegion))
}
