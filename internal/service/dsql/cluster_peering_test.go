// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package dsql_test

import (
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/service/dsql"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccDSQLClusterPeering_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var cluster dsql.GetClusterOutput
	resourceName := "aws_dsql_cluster_peering.test"
	resourceName2 := "aws_dsql_cluster_peering.test1"

	resourceNameCluster := "aws_dsql_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			// Because dsql is in preview, we need to skip the precheck
			// acctest.PreCheckPartitionHasService(t, names.DSQLEndpointID)
			// PreCheck for the region configuration as long as DSQL is in preview
			acctest.PreCheckRegion(t, "us-east-1", "us-east-2")          //lintignore:AWSAT003
			acctest.PreCheckAlternateRegion(t, "us-east-2", "us-east-1") //lintignore:AWSAT003
			acctest.PreCheckThirdRegion(t, "us-west-2")                  //lintignore:AWSAT003
			testAccPreCheck(ctx, t)
			acctest.PreCheckMultipleRegion(t, 2)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.DSQLServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesMultipleRegions(ctx, t, 2),
		// CheckDestroy:             testAccCheckClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterPeeringConfig_basicPrep(acctest.ThirdRegion()),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceNameCluster, &cluster),
				),
			},
			{
				Config: testAccClusterPeeringConfig_basic(acctest.ThirdRegion()),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceNameCluster, &cluster),
					acctest.MatchResourceAttrRegionalARNRegion(ctx, resourceName, "clusters.0", "dsql", acctest.AlternateRegion(), regexache.MustCompile(`cluster/.+$`)),
					acctest.MatchResourceAttrRegionalARNRegion(ctx, resourceName2, "clusters.0", "dsql", acctest.Region(), regexache.MustCompile(`cluster/.+$`)),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, names.AttrIdentifier),
				ImportStateVerifyIdentifierAttribute: names.AttrIdentifier,
				ImportState:                          true,
				ImportStateVerify:                    true,
			},
		},
	})
}

func testAccClusterPeeringConfig_basicPrep(witnessRegion string) string {
	return acctest.ConfigCompose(acctest.ConfigMultipleRegionProvider(2), fmt.Sprintf(`
resource "aws_dsql_cluster" "test" {
  deletion_protection_enabled = false
  multi_region_properties {
    witness_region = "%[1]s"
  }
}

resource "aws_dsql_cluster" "test1" {
  provider = "awsalternate"

  deletion_protection_enabled = false
  multi_region_properties {
    witness_region = "%[1]s"
  }
}
`, witnessRegion))
}

func testAccClusterPeeringConfig_basic(witnessRegion string) string {
	return acctest.ConfigCompose(acctest.ConfigMultipleRegionProvider(2), fmt.Sprintf(`
resource "aws_dsql_cluster" "test" {
  deletion_protection_enabled = false
  multi_region_properties {
    witness_region = "%[1]s"
  }
}

resource "aws_dsql_cluster_peering" "test" {
  identifier     = aws_dsql_cluster.test.identifier
  clusters       = [aws_dsql_cluster.test1.arn]
  witness_region = "%[1]s"
}

resource "aws_dsql_cluster" "test1" {
  provider = "awsalternate"

  deletion_protection_enabled = false
  multi_region_properties {
    witness_region = "%[1]s"
  }
}

resource "aws_dsql_cluster_peering" "test1" {
  provider = "awsalternate"

  identifier     = aws_dsql_cluster.test1.identifier
  clusters       = [aws_dsql_cluster.test.arn]
  witness_region = "%[1]s"
}
`, witnessRegion))
}
