// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2_test

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/YakDriver/regexache"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccEC2CapacityBlockReservation_basic(t *testing.T) {
	ctx := acctest.Context(t)
	key := "RUN_EC2_CAPACITY_BLOCK_RESERVATION_TESTS"
	vifId := os.Getenv(key)
	if vifId != acctest.CtTrue {
		t.Skipf("Environment variable %s is not set to true", key)
	}

	var reservation awstypes.CapacityReservation
	resourceName := "aws_ec2_capacity_block_reservation.test"
	dataSourceName := "data.aws_ec2_capacity_block_offering.test"
	startDate := time.Now().UTC().Add(25 * time.Hour).Format(time.RFC3339)
	endDate := time.Now().UTC().Add(720 * time.Hour).Format(time.RFC3339)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             acctest.CheckDestroyNoop,
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		Steps: []resource.TestStep{
			{
				Config: testAccCapacityBlockReservationConfig_basic(startDate, endDate),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCapacityBlockReservationExists(ctx, resourceName, &reservation),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "ec2", regexache.MustCompile(`capacity-reservation/cr-:.+`)),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrAvailabilityZone, resourceName, names.AttrAvailabilityZone),
					resource.TestCheckResourceAttrPair(dataSourceName, "capacity_block_offering_id", resourceName, "capacity_block_offering_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "start_date", resourceName, "start_date"),
					resource.TestCheckResourceAttrPair(dataSourceName, "end_date", resourceName, "end_date"),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrInstanceCount, resourceName, names.AttrInstanceCount),
					resource.TestCheckResourceAttrPair(dataSourceName, "instance_platform", resourceName, "instance_platform"),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrInstanceType, resourceName, names.AttrInstanceType),
					resource.TestCheckResourceAttrPair(dataSourceName, "tenancy", resourceName, "tenancy"),
				),
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
  instance_type     = "p4d.24xlarge"
  capacity_duration = 24
  instance_count    = 1
  start_date        = %[1]q
  end_date          = %[2]q
}

resource "aws_ec2_capacity_block_reservation" "test" {
  capacity_block_offering_id = data.aws_ec2_capacity_block_offering.test.id
  instance_platform          = "Linux/UNIX"
  tags = {
    "Environment" = "dev"
  }
}
`, startDate, endDate)
}
