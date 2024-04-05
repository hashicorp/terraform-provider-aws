// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package redshift_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/redshift"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccRedshiftOrderableClusterDataSource_clusterType(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_redshift_orderable_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccOrderableClusterPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RedshiftServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             nil,
		Steps: []resource.TestStep{
			{
				Config: testAccOrderableClusterDataSourceConfig_type("multi-node"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "cluster_type", "multi-node"),
				),
			},
		},
	})
}

func TestAccRedshiftOrderableClusterDataSource_clusterVersion(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_redshift_orderable_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccOrderableClusterPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RedshiftServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             nil,
		Steps: []resource.TestStep{
			{
				Config: testAccOrderableClusterDataSourceConfig_version("1.0"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "cluster_version", "1.0"),
				),
			},
		},
	})
}

func TestAccRedshiftOrderableClusterDataSource_nodeType(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_redshift_orderable_cluster.test"
	nodeType := "dc2.8xlarge"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccOrderableClusterPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RedshiftServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             nil,
		Steps: []resource.TestStep{
			{
				Config: testAccOrderableClusterDataSourceConfig_nodeType(nodeType),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "node_type", nodeType),
				),
			},
		},
	})
}

func TestAccRedshiftOrderableClusterDataSource_preferredNodeTypes(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_redshift_orderable_cluster.test"
	preferredNodeType := "dc2.8xlarge"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccOrderableClusterPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RedshiftServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             nil,
		Steps: []resource.TestStep{
			{
				Config: testAccOrderableClusterDataSourceConfig_preferredNodeTypes(preferredNodeType),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "node_type", preferredNodeType),
				),
			},
		},
	})
}

func testAccOrderableClusterPreCheck(ctx context.Context, t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).RedshiftConn(ctx)

	input := &redshift.DescribeOrderableClusterOptionsInput{
		MaxRecords: aws.Int64(20),
	}

	_, err := conn.DescribeOrderableClusterOptionsWithContext(ctx, input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccOrderableClusterDataSourceConfig_type(clusterType string) string {
	return fmt.Sprintf(`
data "aws_redshift_orderable_cluster" "test" {
  cluster_type         = %[1]q
  preferred_node_types = ["dc2.large", "ds2.xlarge"]
}
`, clusterType)
}

func testAccOrderableClusterDataSourceConfig_version(clusterVersion string) string {
	return fmt.Sprintf(`
data "aws_redshift_orderable_cluster" "test" {
  cluster_version      = %[1]q
  preferred_node_types = ["dc2.8xlarge", "ds2.8xlarge"]
}
`, clusterVersion)
}

func testAccOrderableClusterDataSourceConfig_nodeType(nodeType string) string {
	return fmt.Sprintf(`
data "aws_redshift_orderable_cluster" "test" {
  node_type            = %[1]q
  preferred_node_types = ["dc2.8xlarge", "ds2.8xlarge"]
}
`, nodeType)
}

func testAccOrderableClusterDataSourceConfig_preferredNodeTypes(preferredNodeType string) string {
	return fmt.Sprintf(`
data "aws_redshift_orderable_cluster" "test" {
  preferred_node_types = [
    "non-existent",
    %[1]q,
    "try-again",
  ]
}
`, preferredNodeType)
}
