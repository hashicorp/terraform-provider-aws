// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2

import (
	"fmt"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccEC2CapacityBlockOffering_basic(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_ec2_capacity_block_offering.test"
	startDate := time.Now().UTC().Add(25 * time.Hour).Format(time.RFC3339)
	endDate := time.Now().UTC().Add(720 * time.Hour).Format(time.RFC3339)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             nil,
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2),
		Steps: []resource.TestStep{
			{
				Config: testAccCapacityBlockOfferingConfig_basic(startDate, endDate),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(dataSourceName, "availability_zone"),
					resource.TestCheckResourceAttr(dataSourceName, "capacity_duration", "24"),
					resource.TestCheckResourceAttr(dataSourceName, "instance_count", "1"),
					resource.TestCheckResourceAttr(dataSourceName, "instance_type", "p4d.24xlarge"),
					resource.TestCheckResourceAttrSet(dataSourceName, "id"),
					resource.TestCheckResourceAttr(dataSourceName, "tenancy", "default"),
					resource.TestCheckResourceAttrSet(dataSourceName, "upfront_fee"),
				),
			},
		},
	})
}

func testAccCapacityBlockOfferingConfig_basic(startDate, endDate string) string {
	return fmt.Sprintf(`
data "aws_ec2_capacity_block_offering" "test" {
  instance_type = "p4d.24xlarge"
  capacity_duration = 24
  instance_count = 1
  start_date = %[1]q
  end_date = %[2]q
}
`, startDate, endDate)
}
