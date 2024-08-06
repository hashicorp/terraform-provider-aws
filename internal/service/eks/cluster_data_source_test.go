// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package eks_test

import (
	"testing"

	"github.com/YakDriver/regexache"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccEKSClusterDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceResourceName := "data.aws_eks_cluster.test"
	resourceName := "aws_eks_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EKSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterDataSourceConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, "access_config.#", dataSourceResourceName, "access_config.#"),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrARN, dataSourceResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(dataSourceResourceName, "certificate_authority.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "certificate_authority.0.data", dataSourceResourceName, "certificate_authority.0.data"),
					resource.TestCheckNoResourceAttr(dataSourceResourceName, "cluster_id"),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrCreatedAt, dataSourceResourceName, names.AttrCreatedAt),
					resource.TestCheckResourceAttr(dataSourceResourceName, "enabled_cluster_log_types.#", acctest.Ct2),
					resource.TestCheckTypeSetElemAttr(dataSourceResourceName, "enabled_cluster_log_types.*", "api"),
					resource.TestCheckTypeSetElemAttr(dataSourceResourceName, "enabled_cluster_log_types.*", "audit"),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrEndpoint, dataSourceResourceName, names.AttrEndpoint),
					resource.TestCheckResourceAttrPair(resourceName, "identity.#", dataSourceResourceName, "identity.#"),
					resource.TestCheckResourceAttrPair(resourceName, "identity.0.oidc.#", dataSourceResourceName, "identity.0.oidc.#"),
					resource.TestCheckResourceAttrPair(resourceName, "identity.0.oidc.0.issuer", dataSourceResourceName, "identity.0.oidc.0.issuer"),
					resource.TestCheckResourceAttrPair(resourceName, "kubernetes_network_config.#", dataSourceResourceName, "kubernetes_network_config.#"),
					resource.TestCheckResourceAttrPair(resourceName, "kubernetes_network_config.0.ip_family", dataSourceResourceName, "kubernetes_network_config.0.ip_family"),
					resource.TestCheckResourceAttrPair(resourceName, "kubernetes_network_config.0.service_ipv4_cidr", dataSourceResourceName, "kubernetes_network_config.0.service_ipv4_cidr"),
					resource.TestCheckResourceAttrPair(resourceName, "kubernetes_network_config.0.service_ipv6_cidr", dataSourceResourceName, "kubernetes_network_config.0.service_ipv6_cidr"),
					resource.TestCheckResourceAttrPair(resourceName, "outpost_config.#", dataSourceResourceName, "outpost_config.#"),
					resource.TestMatchResourceAttr(dataSourceResourceName, "platform_version", regexache.MustCompile(`^eks\.\d+$`)),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrRoleARN, dataSourceResourceName, names.AttrRoleARN),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrStatus, dataSourceResourceName, names.AttrStatus),
					resource.TestCheckResourceAttrPair(resourceName, acctest.CtTagsPercent, dataSourceResourceName, acctest.CtTagsPercent),
					resource.TestCheckResourceAttr(resourceName, "upgrade_policy.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "upgrade_policy.0.support_type", "EXTENDED"),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrVersion, dataSourceResourceName, names.AttrVersion),
					resource.TestCheckResourceAttr(dataSourceResourceName, "vpc_config.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "vpc_config.0.cluster_security_group_id", dataSourceResourceName, "vpc_config.0.cluster_security_group_id"),
					resource.TestCheckResourceAttrPair(resourceName, "vpc_config.0.endpoint_private_access", dataSourceResourceName, "vpc_config.0.endpoint_private_access"),
					resource.TestCheckResourceAttrPair(resourceName, "vpc_config.0.endpoint_public_access", dataSourceResourceName, "vpc_config.0.endpoint_public_access"),
					resource.TestCheckResourceAttrPair(resourceName, "vpc_config.0.security_group_ids.#", dataSourceResourceName, "vpc_config.0.security_group_ids.#"),
					resource.TestCheckResourceAttrPair(resourceName, "vpc_config.0.subnet_ids.#", dataSourceResourceName, "vpc_config.0.subnet_ids.#"),
					resource.TestCheckResourceAttrPair(resourceName, "vpc_config.0.public_access_cidrs.#", dataSourceResourceName, "vpc_config.0.public_access_cidrs.#"),
					resource.TestCheckResourceAttrPair(resourceName, "vpc_config.0.vpc_id", dataSourceResourceName, "vpc_config.0.vpc_id"),
				),
			},
		},
	})
}

func TestAccEKSClusterDataSource_outpost(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceResourceName := "data.aws_eks_cluster.test"
	resourceName := "aws_eks_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckOutpostsOutposts(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EKSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterDataSourceConfig_outpost(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, names.AttrARN, dataSourceResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(dataSourceResourceName, "certificate_authority.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "certificate_authority.0.data", dataSourceResourceName, "certificate_authority.0.data"),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrCreatedAt, dataSourceResourceName, names.AttrCreatedAt),
					resource.TestCheckResourceAttr(dataSourceResourceName, "enabled_cluster_log_types.#", acctest.Ct0),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrEndpoint, dataSourceResourceName, names.AttrEndpoint),
					resource.TestCheckResourceAttr(dataSourceResourceName, "identity.#", acctest.Ct0),
					resource.TestCheckResourceAttrPair(resourceName, "kubernetes_network_config.#", dataSourceResourceName, "kubernetes_network_config.#"),
					resource.TestCheckResourceAttrPair(resourceName, "kubernetes_network_config.0.ip_family", dataSourceResourceName, "kubernetes_network_config.0.ip_family"),
					resource.TestCheckResourceAttrPair(resourceName, "kubernetes_network_config.0.service_ipv4_cidr", dataSourceResourceName, "kubernetes_network_config.0.service_ipv4_cidr"),
					resource.TestCheckResourceAttrPair(resourceName, "kubernetes_network_config.0.service_ipv6_cidr", dataSourceResourceName, "kubernetes_network_config.0.service_ipv6_cidr"),
					resource.TestMatchResourceAttr(dataSourceResourceName, "platform_version", regexache.MustCompile(`^eks-local-outposts\.\d+$`)),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrRoleARN, dataSourceResourceName, names.AttrRoleARN),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrStatus, dataSourceResourceName, names.AttrStatus),
					resource.TestCheckResourceAttrPair(resourceName, acctest.CtTagsPercent, dataSourceResourceName, acctest.CtTagsPercent),
					resource.TestCheckResourceAttr(resourceName, "upgrade_policy.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "upgrade_policy.0.support_type", "EXTENDED"),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrVersion, dataSourceResourceName, names.AttrVersion),
					resource.TestCheckResourceAttr(dataSourceResourceName, "vpc_config.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "vpc_config.0.cluster_security_group_id", dataSourceResourceName, "vpc_config.0.cluster_security_group_id"),
					resource.TestCheckResourceAttrPair(resourceName, "vpc_config.0.endpoint_private_access", dataSourceResourceName, "vpc_config.0.endpoint_private_access"),
					resource.TestCheckResourceAttrPair(resourceName, "vpc_config.0.endpoint_public_access", dataSourceResourceName, "vpc_config.0.endpoint_public_access"),
					resource.TestCheckResourceAttrPair(resourceName, "vpc_config.0.security_group_ids.#", dataSourceResourceName, "vpc_config.0.security_group_ids.#"),
					resource.TestCheckResourceAttrPair(resourceName, "vpc_config.0.subnet_ids.#", dataSourceResourceName, "vpc_config.0.subnet_ids.#"),
					resource.TestCheckResourceAttrPair(resourceName, "vpc_config.0.public_access_cidrs.#", dataSourceResourceName, "vpc_config.0.public_access_cidrs.#"),
					resource.TestCheckResourceAttrPair(resourceName, "vpc_config.0.vpc_id", dataSourceResourceName, "vpc_config.0.vpc_id"),
					resource.TestCheckResourceAttrPair(resourceName, "outpost_config.0.control_plane_instance_type", dataSourceResourceName, "outpost_config.0.control_plane_instance_type"),
					resource.TestCheckResourceAttrPair(resourceName, "outpost_config.0.control_plane_placement.0.group_name", dataSourceResourceName, "outpost_config.0.control_plane_placement.0.group_name"),
					resource.TestCheckResourceAttrPair(resourceName, "outpost_config.0.outpost_arns.#", dataSourceResourceName, "outpost_config.0.outpost_arns.#"),
				),
			},
		},
	})
}

func testAccClusterDataSourceConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccClusterConfig_logging(rName, []string{"api", "audit"}), `
data "aws_eks_cluster" "test" {
  name = aws_eks_cluster.test.name
}
`)
}

func testAccClusterDataSourceConfig_outpost(rName string) string {
	return acctest.ConfigCompose(testAccClusterConfig_outpostPlacement(rName), `
data "aws_eks_cluster" "test" {
  name = aws_eks_cluster.test.name
}
`)
}
