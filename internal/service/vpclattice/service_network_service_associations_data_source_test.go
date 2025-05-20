// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package vpclattice_test

import (
	"fmt"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccVPCLatticeServiceNetworkServiceAssociationsDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)

	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.VPCLatticeEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.VPCLatticeServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceNetworkServiceAssociationsDataSourceConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckTypeSetElemAttrPair(
						"data.aws_vpclattice_service_network_service_associations.test_sn_1",
						"associations.*.service_network_arn",
						"aws_vpclattice_service_network.test_sn_1",
						names.AttrARN),
					resource.TestCheckTypeSetElemAttrPair(
						"data.aws_vpclattice_service_network_service_associations.test_sn_2",
						"associations.*.service_network_arn",
						"aws_vpclattice_service_network.test_sn_2",
						names.AttrARN),
					resource.TestCheckTypeSetElemAttrPair(
						"data.aws_vpclattice_service_network_service_associations.test_svc_1",
						"associations.*.service_network_arn",
						"aws_vpclattice_service_network.test_sn_1",
						names.AttrARN),
					resource.TestCheckTypeSetElemAttrPair(
						"data.aws_vpclattice_service_network_service_associations.test_svc_2",
						"associations.*.service_network_arn",
						"aws_vpclattice_service_network.test_sn_1",
						names.AttrARN),
					resource.TestCheckTypeSetElemAttrPair(
						"data.aws_vpclattice_service_network_service_associations.test_svc_1",
						"associations.*.service_network_arn",
						"aws_vpclattice_service_network.test_sn_2",
						names.AttrARN),
					resource.TestCheckTypeSetElemAttrPair(
						"data.aws_vpclattice_service_network_service_associations.test_svc_3",
						"associations.*.service_network_arn",
						"aws_vpclattice_service_network.test_sn_2",
						names.AttrARN),
					resource.TestCheckTypeSetElemAttrPair(
						"data.aws_vpclattice_service_network_service_associations.test_sn_1",
						"associations.*.service_arn",
						"aws_vpclattice_service.test_svc_1",
						names.AttrARN),
					resource.TestCheckTypeSetElemAttrPair(
						"data.aws_vpclattice_service_network_service_associations.test_sn_1",
						"associations.*.service_arn",
						"aws_vpclattice_service.test_svc_2",
						names.AttrARN),
					resource.TestCheckTypeSetElemAttrPair(
						"data.aws_vpclattice_service_network_service_associations.test_sn_2",
						"associations.*.service_arn",
						"aws_vpclattice_service.test_svc_1",
						names.AttrARN),
					resource.TestCheckTypeSetElemAttrPair(
						"data.aws_vpclattice_service_network_service_associations.test_sn_2",
						"associations.*.service_arn",
						"aws_vpclattice_service.test_svc_3",
						names.AttrARN),
				),
			},
		},
	})
}

func testAccServiceNetworkServiceAssociationsDataSourceConfig_basic(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

data "aws_caller_identity" "current" {}

resource "aws_vpclattice_service_network" "test_sn_1" {
  name = "%[1]s-sn-1"
}

resource "aws_vpclattice_service_network" "test_sn_2" {
  name = "%[1]s-sn-2"
}

resource "aws_vpclattice_service" "test_svc_1" {
  name               = "%[1]s-svc-1"
  custom_domain_name = "service1.example.com"
}

resource "aws_vpclattice_service" "test_svc_2" {
  name               = "%[1]s-svc-2"
  custom_domain_name = "service2.example.com"
}

resource "aws_vpclattice_service" "test_svc_3" {
  name               = "%[1]s-svc-3"
  custom_domain_name = "service3.example.com"
}

resource "aws_vpclattice_service_network_service_association" "test_sn_1_svc_1" {
  service_identifier         = aws_vpclattice_service.test_svc_1.id
  service_network_identifier = aws_vpclattice_service_network.test_sn_1.id
}

resource "aws_vpclattice_service_network_service_association" "test_sn_1_svc_2" {
  service_identifier         = aws_vpclattice_service.test_svc_2.id
  service_network_identifier = aws_vpclattice_service_network.test_sn_1.id
}

resource "aws_vpclattice_service_network_service_association" "test_sn_2_svc_1" {
  service_identifier         = aws_vpclattice_service.test_svc_1.id
  service_network_identifier = aws_vpclattice_service_network.test_sn_2.id
}

resource "aws_vpclattice_service_network_service_association" "test_sn_2_svc_3" {
  service_identifier         = aws_vpclattice_service.test_svc_3.id
  service_network_identifier = aws_vpclattice_service_network.test_sn_2.id
}

data "aws_vpclattice_service_network_service_associations" "test_sn_1" {
  service_network_identifier = aws_vpclattice_service_network.test_sn_1.id
  depends_on                 = [aws_vpclattice_service_network_service_association.test_sn_1_svc_1, aws_vpclattice_service_network_service_association.test_sn_1_svc_2]
}

data "aws_vpclattice_service_network_service_associations" "test_sn_2" {
  service_network_identifier = aws_vpclattice_service_network.test_sn_2.arn
  depends_on                 = [aws_vpclattice_service_network_service_association.test_sn_2_svc_1, aws_vpclattice_service_network_service_association.test_sn_2_svc_3]
}

data "aws_vpclattice_service_network_service_associations" "test_svc_1" {
  service_identifier = aws_vpclattice_service.test_svc_1.id
  depends_on         = [aws_vpclattice_service_network_service_association.test_sn_1_svc_1, aws_vpclattice_service_network_service_association.test_sn_2_svc_1]
}

data "aws_vpclattice_service_network_service_associations" "test_svc_2" {
  service_identifier = aws_vpclattice_service.test_svc_2.id
  depends_on         = [aws_vpclattice_service_network_service_association.test_sn_1_svc_2]
}

data "aws_vpclattice_service_network_service_associations" "test_svc_3" {
  service_identifier = aws_vpclattice_service.test_svc_3.arn
  depends_on         = [aws_vpclattice_service_network_service_association.test_sn_2_svc_3]
}

`, rName)
}
