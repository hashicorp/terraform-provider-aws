// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package dsql_test

import (
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
			testAccPreCheck(ctx, t)
			acctest.PreCheckMultipleRegion(t, 2)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.DSQLServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesMultipleRegions(ctx, t, 2),
		// CheckDestroy:             testAccCheckClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterPeeringConfig_basicPrep(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceNameCluster, &cluster),
				),
			},
			{
				Config: testAccClusterPeeringConfig_basic(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceNameCluster, &cluster),
					acctest.MatchResourceAttrRegionalARNRegion(ctx, resourceName, "clusters.0", "dsql", "us-east-2", regexache.MustCompile(`cluster/.+$`)),
					acctest.MatchResourceAttrRegionalARNRegion(ctx, resourceName2, "clusters.0", "dsql", "us-east-1", regexache.MustCompile(`cluster/.+$`)),
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

func testAccClusterPeeringConfig_basicPrep() string {
	return acctest.ConfigCompose(acctest.ConfigMultipleRegionProvider(2), `
resource "aws_dsql_cluster" "test" {
  deletion_protection_enabled = false
  multi_region_properties {
    witness_region = "us-west-2"
  }
}

resource "aws_dsql_cluster" "test1" {
  provider = "awsalternate"

  deletion_protection_enabled = false
  multi_region_properties {
    witness_region = "us-west-2"
  }
}
`)
}

func testAccClusterPeeringConfig_basic() string {
	return acctest.ConfigCompose(acctest.ConfigMultipleRegionProvider(2), `
resource "aws_dsql_cluster" "test" {
  deletion_protection_enabled = false
  multi_region_properties {
    witness_region = "us-west-2"
  }
}

resource "aws_dsql_cluster_peering" "test" {
  identifier     = aws_dsql_cluster.test.identifier
  clusters       = [aws_dsql_cluster.test1.arn]
  witness_region = "us-west-2"
}

resource "aws_dsql_cluster" "test1" {
  provider = "awsalternate"

  deletion_protection_enabled = false
  multi_region_properties {
    witness_region = "us-west-2"
  }
}

resource "aws_dsql_cluster_peering" "test1" {
  provider = "awsalternate"

  identifier     = aws_dsql_cluster.test1.identifier
  clusters       = [aws_dsql_cluster.test.arn]
  witness_region = "us-west-2"
}
`)
}
