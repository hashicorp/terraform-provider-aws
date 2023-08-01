// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package vpclattice_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/vpclattice"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccVPCLatticeServiceDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)

	var service vpclattice.GetServiceOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_vpclattice_service.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.VPCLatticeEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.VPCLatticeEndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckServiceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccServiceDataSourceConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists(ctx, dataSourceName, &service),
					resource.TestCheckResourceAttr(dataSourceName, "name", rName),
					resource.TestCheckResourceAttr(dataSourceName, "auth_type", "NONE"),
					acctest.MatchResourceAttrRegionalARN(dataSourceName, "arn", "vpc-lattice", regexp.MustCompile(`service/.+$`)),
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
