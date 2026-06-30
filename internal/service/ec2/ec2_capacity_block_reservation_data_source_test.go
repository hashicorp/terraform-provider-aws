// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package ec2_test

import (
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// Capacity Blocks must be purchased upfront with real money and only exist
// for limited GPU/ML instance types, so the acceptance tests below assume an
// existing pre-provisioned Capacity Block reservation discoverable via the
// CAPACITY_BLOCK_RESERVATION_ID environment variable. Tests are skipped when
// the variable is not set.

func TestAccEC2CapacityBlockReservationDataSource_id(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_ec2_capacity_block_reservation.test"

	reservationID := acctest.SkipIfEnvVarNotSet(t, "TF_AWS_CAPACITY_BLOCK_RESERVATION_ID")

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccCapacityBlockReservationDataSourceConfig_id(reservationID),
				Check: resource.ComposeAggregateTestCheckFunc(
					acctest.MatchResourceAttrRegionalARN(ctx, dataSourceName, names.AttrARN, "ec2", regexache.MustCompile(`capacity-reservation/cr-[a-f0-9]+$`)),
					resource.TestCheckResourceAttrSet(dataSourceName, names.AttrAvailabilityZone),
					resource.TestCheckResourceAttrSet(dataSourceName, names.AttrInstanceCount),
					resource.TestCheckResourceAttrSet(dataSourceName, names.AttrInstanceType),
					resource.TestCheckResourceAttrSet(dataSourceName, "instance_platform"),
					resource.TestCheckResourceAttrSet(dataSourceName, "start_date"),
					resource.TestCheckResourceAttrSet(dataSourceName, "end_date"),
					resource.TestCheckResourceAttr(dataSourceName, "reservation_type", "capacity-block"),
				),
			},
		},
	})
}

func testAccCapacityBlockReservationDataSourceConfig_id(reservationID string) string {
	return fmt.Sprintf(`
data "aws_ec2_capacity_block_reservation" "test" {
  id = %[1]q
}
`, reservationID)
}
