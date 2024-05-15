// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package emrcontainers_test

import (
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccEMRContainersVirtualClusterDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
	dataSourceResourceName := "data.aws_emrcontainers_virtual_cluster.test"
	resourceName := "aws_emrcontainers_virtual_cluster.test"
	testExternalProviders := map[string]resource.ExternalProvider{
		"kubernetes": {
			Source:            "hashicorp/kubernetes",
			VersionConstraint: "~> 2.3",
		},
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckIAMServiceLinkedRole(ctx, t, "/aws-service-role/emr-containers.amazonaws.com")
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EKSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		ExternalProviders:        testExternalProviders,
		CheckDestroy:             testAccCheckVirtualClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVirtualClusterDataSourceConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, names.AttrARN, dataSourceResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(dataSourceResourceName, "container_provider.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "container_provider.0.id", dataSourceResourceName, "container_provider.0.id"),
					resource.TestCheckResourceAttr(dataSourceResourceName, "container_provider.0.info.#", acctest.Ct1),
					resource.TestCheckResourceAttr(dataSourceResourceName, "container_provider.0.info.0.eks_info.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "container_provider.0.info.0.eks_info.0.namespace", dataSourceResourceName, "container_provider.0.info.0.eks_info.0.namespace"),
					resource.TestCheckResourceAttrPair(resourceName, "container_provider.0.type", dataSourceResourceName, "container_provider.0.type"),
					resource.TestCheckResourceAttrSet(dataSourceResourceName, names.AttrCreatedAt),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrID, dataSourceResourceName, names.AttrID),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrName, dataSourceResourceName, names.AttrName),
					resource.TestCheckResourceAttrSet(dataSourceResourceName, names.AttrState),
					resource.TestCheckResourceAttrPair(resourceName, acctest.CtTagsPercent, dataSourceResourceName, acctest.CtTagsPercent),
				),
			},
		},
	})
}

func testAccVirtualClusterDataSourceConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccVirtualClusterConfig_basic(rName), `
data "aws_emrcontainers_virtual_cluster" "test" {
  virtual_cluster_id = aws_emrcontainers_virtual_cluster.test.id
}
`)
}
