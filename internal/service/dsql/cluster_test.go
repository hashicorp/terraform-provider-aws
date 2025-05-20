// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package dsql_test

import (
	"context"
	"errors"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/service/dsql"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tfdsql "github.com/hashicorp/terraform-provider-aws/internal/service/dsql"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccDSQLCluster_basic(t *testing.T) {
	ctx := acctest.Context(t)

	var cluster dsql.GetClusterOutput
	resourceName := "aws_dsql_cluster.test"

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
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.DSQLServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_basic(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &cluster),
					resource.TestCheckResourceAttr(resourceName, "deletion_protection_enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "0"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsAllPercent, "0"),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "dsql", regexache.MustCompile(`cluster/.+$`)),
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

func TestAccDSQLCluster_disappears(t *testing.T) {
	ctx := acctest.Context(t)

	var cluster dsql.GetClusterOutput
	resourceName := "aws_dsql_cluster.test"

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
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.DSQLServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_basic(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &cluster),
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tfdsql.ResourceCluster, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckClusterDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).DSQLClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_dsql_cluster" {
				continue
			}

			_, err := tfdsql.FindClusterByID(ctx, conn, rs.Primary.Attributes[names.AttrIdentifier])
			if tfresource.NotFound(err) {
				return nil
			}
			if err != nil {
				return create.Error(names.DSQL, create.ErrActionCheckingDestroyed, tfdsql.ResNameCluster, rs.Primary.Attributes[names.AttrIdentifier], err)
			}

			return create.Error(names.DSQL, create.ErrActionCheckingDestroyed, tfdsql.ResNameCluster, rs.Primary.Attributes[names.AttrIdentifier], errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccCheckClusterExists(ctx context.Context, name string, cluster *dsql.GetClusterOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.DSQL, create.ErrActionCheckingExistence, tfdsql.ResNameCluster, name, errors.New("not found"))
		}

		if rs.Primary.Attributes[names.AttrIdentifier] == "" {
			return create.Error(names.DSQL, create.ErrActionCheckingExistence, tfdsql.ResNameCluster, name, errors.New("not set"))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).DSQLClient(ctx)

		resp, err := tfdsql.FindClusterByID(ctx, conn, rs.Primary.Attributes[names.AttrIdentifier])
		if err != nil {
			return create.Error(names.DSQL, create.ErrActionCheckingExistence, tfdsql.ResNameCluster, rs.Primary.Attributes[names.AttrIdentifier], err)
		}

		*cluster = *resp

		return nil
	}
}

func testAccPreCheck(ctx context.Context, t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).DSQLClient(ctx)

	input := dsql.ListClustersInput{}
	_, err := conn.ListClusters(ctx, &input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}
	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccClusterConfig_basic() string {
	return `
resource "aws_dsql_cluster" "test" {
  deletion_protection_enabled = false
}
`
}
