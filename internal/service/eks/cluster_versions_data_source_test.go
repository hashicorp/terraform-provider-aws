// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package eks_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccEKSClusterVersionsDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)

	dataSourceName := "data.aws_eks_cluster_versions.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EKSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterVersionsDataSourceConfig_basic(),
				Check: resource.ComposeAggregateTestCheckFunc(
					acctest.CheckResourceAttrGreaterThanValue(dataSourceName, "cluster_versions.#", 0),
					acctest.CheckResourceAttrContains(dataSourceName, "cluster_versions.0.default_version", acctest.CtTrue),
				),
			},
		},
	})
}

func TestAccEKSClusterVersionsDataSource_clusterType(t *testing.T) {
	ctx := acctest.Context(t)

	dataSourceName := "data.aws_eks_cluster_versions.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EKSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterVersionsDataSourceConfig_clusterType(),
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckResourceAttrGreaterThanValue(dataSourceName, "cluster_versions.#", 0),
				),
			},
		},
	})
}

func TestAccEKSClusterVersionsDataSource_defaultOnly(t *testing.T) {
	ctx := acctest.Context(t)

	dataSourceName := "data.aws_eks_cluster_versions.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EKSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterVersionsDataSourceConfig_defaultOnly(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "cluster_versions.#", "1"),
					acctest.CheckResourceAttrContains(dataSourceName, "cluster_versions.0.default_version", acctest.CtTrue),
				),
			},
		},
	})
}

func TestAccEKSClusterVersionsDataSource_versionStatus(t *testing.T) {
	ctx := acctest.Context(t)

	dataSourceName := "data.aws_eks_cluster_versions.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EKSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterVersionsDataSourceConfig_versionStatus(),
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckResourceAttrGreaterThanValue(dataSourceName, "cluster_versions.#", 0),
					acctest.CheckResourceAttrContains(dataSourceName, "cluster_versions.0.version_status", "STANDARD_SUPPORT"),
				),
			},
		},
	})
}

func testAccClusterVersionsDataSourceConfig_basic() string {
	return `
data "aws_eks_cluster_versions" "test" {}
`
}

func testAccClusterVersionsDataSourceConfig_clusterType() string {
	return `
data "aws_eks_cluster_versions" "test" {
  cluster_type = "eks"
}
`
}

func testAccClusterVersionsDataSourceConfig_defaultOnly() string {
	return `
data "aws_eks_cluster_versions" "test" {
  default_only = true
}
`
}

func testAccClusterVersionsDataSourceConfig_versionStatus() string {
	return `
data "aws_eks_cluster_versions" "test" {
  version_status = "STANDARD_SUPPORT"
}
`
}
