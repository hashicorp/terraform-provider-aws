// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	awstypes "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// NOTE on pricing
//
// Due to pricing, Capacity Block Reservation tests are skipped by default.
// They can be run by setting the environment variable TF_AWS_RUN_EC2_CAPACITY_BLOCK_RESERVATION_TESTS.
//
// https://aws.amazon.com/ec2/capacityblocks/pricing/
//
// As of 2025-12-04, the instance types available in more than one region have the following hourly costs:
// p6-b200.48xlarge: $74.88  USD
// p5.48xlarge:      $31.464 USD
// p5.4xlarge:        $3.933 USD
// p5e.48xlarge:     $34.608 USD

func TestAccEC2CapacityBlockReservation_basic(t *testing.T) {
	ctx := acctest.Context(t)

	acctest.SkipIfEnvVarNotSet(t, "TF_AWS_RUN_EC2_CAPACITY_BLOCK_RESERVATION_TESTS")

	var reservation awstypes.CapacityReservation
	resourceName := "aws_ec2_capacity_block_reservation.test"
	dataSourceName := "data.aws_ec2_capacity_block_offering.test"
	startDate := time.Now().UTC().Format(time.RFC3339)
	endDate := time.Now().UTC().Add(720 * time.Hour).Format(time.RFC3339)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             acctest.CheckDestroyNoop,
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		Steps: []resource.TestStep{
			{
				Config: testAccCapacityBlockReservationConfig_basic(startDate, endDate),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCapacityBlockReservationExists(ctx, resourceName, &reservation),
					acctest.CheckResourceAttrRegionalARNFormat(ctx, resourceName, names.AttrARN, "ec2", "capacity-reservation/{id}"),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrAvailabilityZone, dataSourceName, names.AttrAvailabilityZone),
					resource.TestCheckResourceAttrPair(resourceName, "capacity_block_offering_id", dataSourceName, "capacity_block_offering_id"),
					// TODO: `start_date` is after `start_date_range`
					// TODO: `end_date` is before `end_date_range`
					// TODO: `instance_count` is not populated until the CBR is active, but requested count is in tag `aws:ec2capacityreservation:incrementalRequestedQuantity`
					// resource.TestCheckResourceAttrPair(resourceName, names.AttrInstanceCount, dataSourceName, names.AttrInstanceCount),
					resource.TestCheckResourceAttr(resourceName, "instance_platform", "Linux/UNIX"),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrInstanceType, dataSourceName, names.AttrInstanceType),
					resource.TestCheckResourceAttrPair(resourceName, "tenancy", dataSourceName, "tenancy"),
				),
				// Because `aws_ec2_capacity_block_offering` is a data source, the `capacity_block_offering_id` changes on each run
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckCapacityBlockReservationExists(ctx context.Context, n string, v *awstypes.CapacityReservation) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Client(ctx)

		output, err := tfec2.FindCapacityReservationByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCapacityBlockReservationConfig_basic(startDate, endDate string) string {
	return fmt.Sprintf(`
data "aws_ec2_capacity_block_offering" "test" {
  instance_type           = "p5.4xlarge"
  capacity_duration_hours = 24
  instance_count          = 1
  start_date_range        = %[1]q
  end_date_range          = %[2]q
}

resource "aws_ec2_capacity_block_reservation" "test" {
  capacity_block_offering_id = data.aws_ec2_capacity_block_offering.test.capacity_block_offering_id
  instance_platform          = "Linux/UNIX"
}
`, startDate, endDate)
}
