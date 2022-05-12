package ec2_test

import (
	"fmt"
	"regexp"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func TestAccEC2CapacityReservation_basic(t *testing.T) {
	var cr ec2.CapacityReservation
	availabilityZonesDataSourceName := "data.aws_availability_zones.available"
	resourceName := "aws_ec2_capacity_reservation.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckCapacityReservation(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckEc2CapacityReservationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEc2CapacityReservationConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEc2CapacityReservationExists(resourceName, &cr),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "ec2", regexp.MustCompile(`capacity-reservation/cr-.+`)),
					resource.TestCheckResourceAttrPair(resourceName, "availability_zone", availabilityZonesDataSourceName, "names.0"),
					resource.TestCheckResourceAttr(resourceName, "ebs_optimized", "false"),
					resource.TestCheckResourceAttr(resourceName, "end_date", ""),
					resource.TestCheckResourceAttr(resourceName, "end_date_type", "unlimited"),
					resource.TestCheckResourceAttr(resourceName, "ephemeral_storage", "false"),
					resource.TestCheckResourceAttr(resourceName, "instance_count", "1"),
					resource.TestCheckResourceAttr(resourceName, "instance_match_criteria", "open"),
					resource.TestCheckResourceAttr(resourceName, "instance_platform", "Linux/UNIX"),
					resource.TestCheckResourceAttr(resourceName, "instance_type", "t2.micro"),
					resource.TestCheckResourceAttr(resourceName, "outpost_arn", ""),
					acctest.CheckResourceAttrAccountID(resourceName, "owner_id"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
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
	var cr ec2.CapacityReservation
	resourceName := "aws_ec2_capacity_reservation.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckCapacityReservation(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckEc2CapacityReservationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEc2CapacityReservationConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEc2CapacityReservationExists(resourceName, &cr),
					acctest.CheckResourceDisappears(acctest.Provider, tfec2.ResourceCapacityReservation(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccEC2CapacityReservation_ebsOptimized(t *testing.T) {
	var cr ec2.CapacityReservation
	resourceName := "aws_ec2_capacity_reservation.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckCapacityReservation(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckEc2CapacityReservationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEc2CapacityReservationConfig_ebsOptimized(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEc2CapacityReservationExists(resourceName, &cr),
					resource.TestCheckResourceAttr(resourceName, "ebs_optimized", "true"),
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
	var cr ec2.CapacityReservation
	endDate1 := time.Now().UTC().Add(1 * time.Hour).Format(time.RFC3339)
	endDate2 := time.Now().UTC().Add(2 * time.Hour).Format(time.RFC3339)
	resourceName := "aws_ec2_capacity_reservation.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckCapacityReservation(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckEc2CapacityReservationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEc2CapacityReservationConfig_endDate(rName, endDate1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEc2CapacityReservationExists(resourceName, &cr),
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
				Config: testAccEc2CapacityReservationConfig_endDate(rName, endDate2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEc2CapacityReservationExists(resourceName, &cr),
					resource.TestCheckResourceAttr(resourceName, "end_date", endDate2),
					resource.TestCheckResourceAttr(resourceName, "end_date_type", "limited"),
				),
			},
		},
	})
}

func TestAccEC2CapacityReservation_endDateType(t *testing.T) {
	var cr ec2.CapacityReservation
	endDate := time.Now().UTC().Add(12 * time.Hour).Format(time.RFC3339)
	resourceName := "aws_ec2_capacity_reservation.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckCapacityReservation(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckEc2CapacityReservationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEc2CapacityReservationConfig_endDateType(rName, "unlimited"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEc2CapacityReservationExists(resourceName, &cr),
					resource.TestCheckResourceAttr(resourceName, "end_date_type", "unlimited"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccEc2CapacityReservationConfig_endDate(rName, endDate),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEc2CapacityReservationExists(resourceName, &cr),
					resource.TestCheckResourceAttr(resourceName, "end_date", endDate),
					resource.TestCheckResourceAttr(resourceName, "end_date_type", "limited"),
				),
			},
			{
				Config: testAccEc2CapacityReservationConfig_endDateType(rName, "unlimited"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEc2CapacityReservationExists(resourceName, &cr),
					resource.TestCheckResourceAttr(resourceName, "end_date_type", "unlimited"),
				),
			},
		},
	})
}

func TestAccEC2CapacityReservation_ephemeralStorage(t *testing.T) {
	var cr ec2.CapacityReservation
	resourceName := "aws_ec2_capacity_reservation.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckCapacityReservation(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckEc2CapacityReservationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEc2CapacityReservationConfig_ephemeralStorage(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEc2CapacityReservationExists(resourceName, &cr),
					resource.TestCheckResourceAttr(resourceName, "ephemeral_storage", "true"),
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
	var cr ec2.CapacityReservation
	resourceName := "aws_ec2_capacity_reservation.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckCapacityReservation(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckEc2CapacityReservationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEc2CapacityReservationConfig_instanceCount(rName, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEc2CapacityReservationExists(resourceName, &cr),
					resource.TestCheckResourceAttr(resourceName, "instance_count", "1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccEc2CapacityReservationConfig_instanceCount(rName, 2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEc2CapacityReservationExists(resourceName, &cr),
					resource.TestCheckResourceAttr(resourceName, "instance_count", "2"),
				),
			},
		},
	})
}

func TestAccEC2CapacityReservation_instanceMatchCriteria(t *testing.T) {
	var cr ec2.CapacityReservation
	resourceName := "aws_ec2_capacity_reservation.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckCapacityReservation(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckEc2CapacityReservationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEc2CapacityReservationConfig_instanceMatchCriteria(rName, "targeted"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEc2CapacityReservationExists(resourceName, &cr),
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
	var cr ec2.CapacityReservation
	resourceName := "aws_ec2_capacity_reservation.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckCapacityReservation(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckEc2CapacityReservationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEc2CapacityReservationConfig_instanceType(rName, "t2.micro"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEc2CapacityReservationExists(resourceName, &cr),
					resource.TestCheckResourceAttr(resourceName, "instance_type", "t2.micro"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccEc2CapacityReservationConfig_instanceType(rName, "t2.small"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEc2CapacityReservationExists(resourceName, &cr),
					resource.TestCheckResourceAttr(resourceName, "instance_type", "t2.small"),
				),
			},
		},
	})
}

func TestAccEC2CapacityReservation_tags(t *testing.T) {
	var cr ec2.CapacityReservation
	resourceName := "aws_ec2_capacity_reservation.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckCapacityReservation(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckEc2CapacityReservationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEc2CapacityReservationConfigTags1("key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEc2CapacityReservationExists(resourceName, &cr),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccEc2CapacityReservationConfigTags2("key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEc2CapacityReservationExists(resourceName, &cr),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccEc2CapacityReservationConfigTags1("key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEc2CapacityReservationExists(resourceName, &cr),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccEC2CapacityReservation_tenancy(t *testing.T) {
	// Error creating EC2 Capacity Reservation: Unsupported: The requested configuration is currently not supported. Please check the documentation for supported configurations.
	acctest.Skip(t, "EC2 Capacity Reservations do not currently support dedicated tenancy.")
	var cr ec2.CapacityReservation
	resourceName := "aws_ec2_capacity_reservation.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckCapacityReservation(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckEc2CapacityReservationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEc2CapacityReservationConfig_tenancy(rName, "dedicated"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEc2CapacityReservationExists(resourceName, &cr),
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

func testAccCheckEc2CapacityReservationExists(n string, v *ec2.CapacityReservation) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No EC2 Capacity Reservation ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn

		output, err := tfec2.FindCapacityReservationByID(conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckEc2CapacityReservationDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_ec2_capacity_reservation" {
			continue
		}

		_, err := tfec2.FindCapacityReservationByID(conn, rs.Primary.ID)

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

func testAccPreCheckCapacityReservation(t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn

	input := &ec2.DescribeCapacityReservationsInput{
		MaxResults: aws.Int64(1),
	}

	_, err := conn.DescribeCapacityReservations(input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

var testAccEc2CapacityReservationConfig = acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptIn(), `
resource "aws_ec2_capacity_reservation" "test" {
  availability_zone = data.aws_availability_zones.available.names[0]
  instance_count    = 1
  instance_platform = "Linux/UNIX"
  instance_type     = "t2.micro"
}
`)

func testAccEc2CapacityReservationConfig_ebsOptimized(rName string, ebsOptimized bool) string {
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

func testAccEc2CapacityReservationConfig_endDate(rName, endDate string) string {
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

func testAccEc2CapacityReservationConfig_endDateType(rName, endDateType string) string {
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

func testAccEc2CapacityReservationConfig_ephemeralStorage(rName string, ephemeralStorage bool) string {
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

func testAccEc2CapacityReservationConfig_instanceCount(rName string, instanceCount int) string {
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

func testAccEc2CapacityReservationConfig_instanceMatchCriteria(rName, instanceMatchCriteria string) string {
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

func testAccEc2CapacityReservationConfig_instanceType(rName, instanceType string) string {
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

func testAccEc2CapacityReservationConfigTags1(tag1Key, tag1Value string) string {
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

func testAccEc2CapacityReservationConfigTags2(tag1Key, tag1Value, tag2Key, tag2Value string) string {
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

func testAccEc2CapacityReservationConfig_tenancy(rName, tenancy string) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptIn(), fmt.Sprintf(`
resource "aws_ec2_capacity_reservation" "test" {
  availability_zone = data.aws_availability_zones.available.names[0]
  instance_count    = 1
  instance_platform = "Linux/UNIX"
  instance_type     = "t2.micro"
  tenancy           = %[2]q

  tags = {
    Name = %[1]q
  }
}
`, rName, tenancy))
}
