// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccEC2CapacityReservation_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var cr awstypes.CapacityReservation
	availabilityZonesDataSourceName := "data.aws_availability_zones.available"
	resourceName := "aws_ec2_capacity_reservation.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckCapacityReservation(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCapacityReservationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCapacityReservationConfig_basic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCapacityReservationExists(ctx, resourceName, &cr),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "ec2", regexache.MustCompile(`capacity-reservation/cr-.+`)),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrAvailabilityZone, availabilityZonesDataSourceName, "names.0"),
					resource.TestCheckResourceAttr(resourceName, "ebs_optimized", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "end_date", ""),
					resource.TestCheckResourceAttr(resourceName, "end_date_type", "unlimited"),
					resource.TestCheckResourceAttr(resourceName, "ephemeral_storage", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, names.AttrInstanceCount, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "instance_match_criteria", "open"),
					resource.TestCheckResourceAttr(resourceName, "instance_platform", "Linux/UNIX"),
					resource.TestCheckResourceAttr(resourceName, names.AttrInstanceType, "t2.micro"),
					resource.TestCheckResourceAttr(resourceName, "outpost_arn", ""),
					acctest.CheckResourceAttrAccountID(resourceName, names.AttrOwnerID),
					resource.TestCheckResourceAttr(resourceName, "placement_group_arn", ""),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "tenancy", "default"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccEC2CapacityReservation_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var cr awstypes.CapacityReservation
	resourceName := "aws_ec2_capacity_reservation.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckCapacityReservation(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCapacityReservationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCapacityReservationConfig_basic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCapacityReservationExists(ctx, resourceName, &cr),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfec2.ResourceCapacityReservation(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccEC2CapacityReservation_ebsOptimized(t *testing.T) {
	ctx := acctest.Context(t)
	var cr awstypes.CapacityReservation
	resourceName := "aws_ec2_capacity_reservation.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckCapacityReservation(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCapacityReservationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCapacityReservationConfig_ebsOptimized(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCapacityReservationExists(ctx, resourceName, &cr),
					resource.TestCheckResourceAttr(resourceName, "ebs_optimized", acctest.CtTrue),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccEC2CapacityReservation_endDate(t *testing.T) {
	ctx := acctest.Context(t)
	var cr awstypes.CapacityReservation
	endDate1 := time.Now().UTC().Add(1 * time.Hour).Format(time.RFC3339)
	endDate2 := time.Now().UTC().Add(2 * time.Hour).Format(time.RFC3339)
	resourceName := "aws_ec2_capacity_reservation.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckCapacityReservation(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCapacityReservationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCapacityReservationConfig_endDate(rName, endDate1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCapacityReservationExists(ctx, resourceName, &cr),
					resource.TestCheckResourceAttr(resourceName, "end_date", endDate1),
					resource.TestCheckResourceAttr(resourceName, "end_date_type", "limited"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccCapacityReservationConfig_endDate(rName, endDate2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCapacityReservationExists(ctx, resourceName, &cr),
					resource.TestCheckResourceAttr(resourceName, "end_date", endDate2),
					resource.TestCheckResourceAttr(resourceName, "end_date_type", "limited"),
				),
			},
		},
	})
}

func TestAccEC2CapacityReservation_endDateType(t *testing.T) {
	ctx := acctest.Context(t)
	var cr awstypes.CapacityReservation
	endDate := time.Now().UTC().Add(12 * time.Hour).Format(time.RFC3339)
	resourceName := "aws_ec2_capacity_reservation.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckCapacityReservation(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCapacityReservationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCapacityReservationConfig_endDateType(rName, "unlimited"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCapacityReservationExists(ctx, resourceName, &cr),
					resource.TestCheckResourceAttr(resourceName, "end_date_type", "unlimited"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccCapacityReservationConfig_endDate(rName, endDate),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCapacityReservationExists(ctx, resourceName, &cr),
					resource.TestCheckResourceAttr(resourceName, "end_date", endDate),
					resource.TestCheckResourceAttr(resourceName, "end_date_type", "limited"),
				),
			},
			{
				Config: testAccCapacityReservationConfig_endDateType(rName, "unlimited"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCapacityReservationExists(ctx, resourceName, &cr),
					resource.TestCheckResourceAttr(resourceName, "end_date_type", "unlimited"),
				),
			},
		},
	})
}

func TestAccEC2CapacityReservation_ephemeralStorage(t *testing.T) {
	ctx := acctest.Context(t)
	var cr awstypes.CapacityReservation
	resourceName := "aws_ec2_capacity_reservation.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckCapacityReservation(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCapacityReservationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCapacityReservationConfig_ephemeralStorage(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCapacityReservationExists(ctx, resourceName, &cr),
					resource.TestCheckResourceAttr(resourceName, "ephemeral_storage", acctest.CtTrue),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccEC2CapacityReservation_instanceCount(t *testing.T) {
	ctx := acctest.Context(t)
	var cr awstypes.CapacityReservation
	resourceName := "aws_ec2_capacity_reservation.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckCapacityReservation(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCapacityReservationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCapacityReservationConfig_instanceCount(rName, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCapacityReservationExists(ctx, resourceName, &cr),
					resource.TestCheckResourceAttr(resourceName, names.AttrInstanceCount, acctest.Ct1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccCapacityReservationConfig_instanceCount(rName, 2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCapacityReservationExists(ctx, resourceName, &cr),
					resource.TestCheckResourceAttr(resourceName, names.AttrInstanceCount, acctest.Ct2),
				),
			},
		},
	})
}

func TestAccEC2CapacityReservation_instanceMatchCriteria(t *testing.T) {
	ctx := acctest.Context(t)
	var cr awstypes.CapacityReservation
	resourceName := "aws_ec2_capacity_reservation.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckCapacityReservation(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCapacityReservationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCapacityReservationConfig_instanceMatchCriteria(rName, "targeted"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCapacityReservationExists(ctx, resourceName, &cr),
					resource.TestCheckResourceAttr(resourceName, "instance_match_criteria", "targeted"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccEC2CapacityReservation_instanceType(t *testing.T) {
	ctx := acctest.Context(t)
	var cr awstypes.CapacityReservation
	resourceName := "aws_ec2_capacity_reservation.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckCapacityReservation(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCapacityReservationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCapacityReservationConfig_instanceType(rName, "t2.micro"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCapacityReservationExists(ctx, resourceName, &cr),
					resource.TestCheckResourceAttr(resourceName, names.AttrInstanceType, "t2.micro"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccCapacityReservationConfig_instanceType(rName, "t2.small"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCapacityReservationExists(ctx, resourceName, &cr),
					resource.TestCheckResourceAttr(resourceName, names.AttrInstanceType, "t2.small"),
				),
			},
		},
	})
}

func TestAccEC2CapacityReservation_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var cr awstypes.CapacityReservation
	resourceName := "aws_ec2_capacity_reservation.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckCapacityReservation(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCapacityReservationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCapacityReservationConfig_tags1(acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCapacityReservationExists(ctx, resourceName, &cr),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccCapacityReservationConfig_tags2(acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCapacityReservationExists(ctx, resourceName, &cr),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccCapacityReservationConfig_tags1(acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCapacityReservationExists(ctx, resourceName, &cr),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func TestAccEC2CapacityReservation_tenancy(t *testing.T) {
	ctx := acctest.Context(t)
	var cr awstypes.CapacityReservation
	resourceName := "aws_ec2_capacity_reservation.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckCapacityReservation(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCapacityReservationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCapacityReservationConfig_tenancy(rName, "dedicated"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCapacityReservationExists(ctx, resourceName, &cr),
					resource.TestCheckResourceAttr(resourceName, "tenancy", "dedicated"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckCapacityReservationExists(ctx context.Context, n string, v *awstypes.CapacityReservation) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No EC2 Capacity Reservation ID is set")
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

func testAccCheckCapacityReservationDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Client(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_ec2_capacity_reservation" {
				continue
			}

			_, err := tfec2.FindCapacityReservationByID(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("EC2 Capacity Reservation %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccPreCheckCapacityReservation(ctx context.Context, t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Client(ctx)

	input := &ec2.DescribeCapacityReservationsInput{
		MaxResults: aws.Int32(1),
	}

	_, err := conn.DescribeCapacityReservations(ctx, input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

var testAccCapacityReservationConfig_basic = acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptIn(), `
resource "aws_ec2_capacity_reservation" "test" {
  availability_zone = data.aws_availability_zones.available.names[0]
  instance_count    = 1
  instance_platform = "Linux/UNIX"
  instance_type     = "t2.micro"
}
`)

func testAccCapacityReservationConfig_ebsOptimized(rName string, ebsOptimized bool) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptIn(), fmt.Sprintf(`
resource "aws_ec2_capacity_reservation" "test" {
  availability_zone = data.aws_availability_zones.available.names[0]
  ebs_optimized     = %[2]t
  instance_count    = 1
  instance_platform = "Linux/UNIX"
  instance_type     = "m4.large"

  tags = {
    Name = %[1]q
  }
}
`, rName, ebsOptimized))
}

func testAccCapacityReservationConfig_endDate(rName, endDate string) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptIn(), fmt.Sprintf(`
resource "aws_ec2_capacity_reservation" "test" {
  availability_zone = data.aws_availability_zones.available.names[0]
  end_date          = %[2]q
  end_date_type     = "limited"
  instance_count    = 1
  instance_platform = "Linux/UNIX"
  instance_type     = "t2.micro"

  tags = {
    Name = %[1]q
  }
}
`, rName, endDate))
}

func testAccCapacityReservationConfig_endDateType(rName, endDateType string) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptIn(), fmt.Sprintf(`
resource "aws_ec2_capacity_reservation" "test" {
  availability_zone = data.aws_availability_zones.available.names[0]
  end_date_type     = %[2]q
  instance_count    = 1
  instance_platform = "Linux/UNIX"
  instance_type     = "t2.micro"

  tags = {
    Name = %[1]q
  }
}
`, rName, endDateType))
}

func testAccCapacityReservationConfig_ephemeralStorage(rName string, ephemeralStorage bool) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptIn(), fmt.Sprintf(`
resource "aws_ec2_capacity_reservation" "test" {
  availability_zone = data.aws_availability_zones.available.names[0]
  ephemeral_storage = %[2]t
  instance_count    = 1
  instance_platform = "Linux/UNIX"
  instance_type     = "m3.medium"

  tags = {
    Name = %[1]q
  }
}
`, rName, ephemeralStorage))
}

func testAccCapacityReservationConfig_instanceCount(rName string, instanceCount int) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptIn(), fmt.Sprintf(`
resource "aws_ec2_capacity_reservation" "test" {
  availability_zone = data.aws_availability_zones.available.names[0]
  instance_count    = %[2]d
  instance_platform = "Linux/UNIX"
  instance_type     = "t2.micro"

  tags = {
    Name = %[1]q
  }
}
`, rName, instanceCount))
}

func testAccCapacityReservationConfig_instanceMatchCriteria(rName, instanceMatchCriteria string) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptIn(), fmt.Sprintf(`
resource "aws_ec2_capacity_reservation" "test" {
  availability_zone       = data.aws_availability_zones.available.names[0]
  instance_count          = 1
  instance_platform       = "Linux/UNIX"
  instance_match_criteria = %[2]q
  instance_type           = "t2.micro"

  tags = {
    Name = %[1]q
  }
}
`, rName, instanceMatchCriteria))
}

func testAccCapacityReservationConfig_instanceType(rName, instanceType string) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptIn(), fmt.Sprintf(`
resource "aws_ec2_capacity_reservation" "test" {
  availability_zone = data.aws_availability_zones.available.names[0]
  instance_count    = 1
  instance_platform = "Linux/UNIX"
  instance_type     = %[2]q

  tags = {
    Name = %[1]q
  }
}
`, rName, instanceType))
}

func testAccCapacityReservationConfig_tags1(tag1Key, tag1Value string) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptIn(), fmt.Sprintf(`
resource "aws_ec2_capacity_reservation" "test" {
  availability_zone = data.aws_availability_zones.available.names[0]
  instance_count    = 1
  instance_platform = "Linux/UNIX"
  instance_type     = "t2.micro"

  tags = {
    %[1]q = %[2]q
  }
}
`, tag1Key, tag1Value))
}

func testAccCapacityReservationConfig_tags2(tag1Key, tag1Value, tag2Key, tag2Value string) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptIn(), fmt.Sprintf(`
resource "aws_ec2_capacity_reservation" "test" {
  availability_zone = data.aws_availability_zones.available.names[0]
  instance_count    = 1
  instance_platform = "Linux/UNIX"
  instance_type     = "t2.micro"

  tags = {
    %[1]q = %[2]q
    %[3]q = %[4]q
  }
}
`, tag1Key, tag1Value, tag2Key, tag2Value))
}

func testAccCapacityReservationConfig_tenancy(rName, tenancy string) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptIn(), fmt.Sprintf(`
resource "aws_ec2_capacity_reservation" "test" {
  availability_zone = data.aws_availability_zones.available.names[1]
  instance_count    = 1
  instance_platform = "Linux/UNIX"
  instance_type     = "a1.4xlarge"
  tenancy           = %[2]q

  tags = {
    Name = %[1]q
  }
}
`, rName, tenancy))
}
