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

func TestAccVPCLatticeServiceDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_vpclattice_service.test"
	dataSourceName := "data.aws_vpclattice_service.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.VPCLatticeEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.VPCLatticeEndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceDataSourceConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, "arn", dataSourceName, "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "auth_type", dataSourceName, "auth_type"),
					resource.TestCheckResourceAttrPair(resourceName, "certificate_arn", dataSourceName, "certificate_arn"),
					resource.TestCheckResourceAttrPair(resourceName, "custom_domain_name", dataSourceName, "custom_domain_name"),
					resource.TestCheckResourceAttrPair(resourceName, "dns_entry.#", dataSourceName, "dns_entry.#"),
					resource.TestCheckResourceAttrPair(resourceName, "name", dataSourceName, "name"),
					resource.TestCheckResourceAttrPair(resourceName, "status", dataSourceName, "status"),
					resource.TestCheckResourceAttrPair(resourceName, "tags.%", dataSourceName, "tags.%"),
				),
			},
		},
	})
}

func TestAccVPCLatticeServiceDataSource_byName(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_vpclattice_service.test"
	dataSourceName := "data.aws_vpclattice_service.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.VPCLatticeEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.VPCLatticeEndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceDataSourceConfig_byName(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, "arn", dataSourceName, "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "auth_type", dataSourceName, "auth_type"),
					resource.TestCheckResourceAttrPair(resourceName, "certificate_arn", dataSourceName, "certificate_arn"),
					resource.TestCheckResourceAttrPair(resourceName, "custom_domain_name", dataSourceName, "custom_domain_name"),
					resource.TestCheckResourceAttrPair(resourceName, "dns_entry.#", dataSourceName, "dns_entry.#"),
					resource.TestCheckResourceAttrPair(resourceName, "name", dataSourceName, "name"),
					resource.TestCheckResourceAttrSet(dataSourceName, "service_identifier"),
					resource.TestCheckResourceAttrPair(resourceName, "status", dataSourceName, "status"),
					resource.TestCheckResourceAttrPair(resourceName, "tags.%", dataSourceName, "tags.%"),
				),
			},
		},
	})
}

func testAccServiceDataSourceConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpclattice_service" "test" {
  name = %[1]q
}

data "aws_vpclattice_service" "test" {
  service_identifier = aws_vpclattice_service.test.id
}
`, rName)
}

func testAccServiceDataSourceConfig_byName(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpclattice_service" "test" {
  name = %[1]q
}

data "aws_vpclattice_service" "test" {
  name = aws_vpclattice_service.test.name
}
`, rName)
}
