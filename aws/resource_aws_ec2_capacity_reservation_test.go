package aws

import (
	"fmt"
	"log"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func init() {
	resource.AddTestSweepers("aws_ec2_capacity_reservation", &resource.Sweeper{
		Name: "aws_ec2_capacity_reservation",
		F:    testSweepEc2CapacityReservations,
	})
}

func testSweepEc2CapacityReservations(region string) error {
	client, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*AWSClient).ec2conn

	resp, err := conn.DescribeCapacityReservations(&ec2.DescribeCapacityReservationsInput{})

	if err != nil {
		return fmt.Errorf("Error retrieving EC2 Capacity Reservations: %s", err)
	}

	if len(resp.CapacityReservations) == 0 {
		log.Print("[DEBUG] No EC2 Capacity Reservations to sweep")
		return nil
	}

	for _, r := range resp.CapacityReservations {
		if aws.StringValue(r.State) != ec2.CapacityReservationStateCancelled && aws.StringValue(r.State) != ec2.CapacityReservationStateExpired {
			id := aws.StringValue(r.CapacityReservationId)

			log.Printf("[INFO] Cancelling EC2 Capacity Reservation EC2 Instance: %s", id)

			opts := &ec2.CancelCapacityReservationInput{
				CapacityReservationId: aws.String(id),
			}

			_, err := conn.CancelCapacityReservation(opts)

			if err != nil {
				log.Printf("[ERROR] Error cancelling EC2 Capacity Reservation (%s): %s", id, err)
			}
		}
	}

	return nil
}

func TestAccAWSEc2CapacityReservation_basic(t *testing.T) {
	var cr ec2.CapacityReservation
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckEc2CapacityReservationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEc2CapacityReservationConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEc2CapacityReservationExists("aws_ec2_capacity_reservation.default", &cr),
					resource.TestCheckResourceAttr("aws_ec2_capacity_reservation.default", "instance_type", "t2.micro"),
				),
			},
			{
				ResourceName:      "aws_ec2_capacity_reservation.default",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSEc2CapacityReservation_endDate(t *testing.T) {
	var cr ec2.CapacityReservation
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckEc2CapacityReservationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEc2CapacityReservationConfig_endDate,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEc2CapacityReservationExists("aws_ec2_capacity_reservation.default", &cr),
					resource.TestCheckResourceAttr("aws_ec2_capacity_reservation.default", "end_date", "2019-10-31T07:39:57Z"),
					resource.TestCheckResourceAttr("aws_ec2_capacity_reservation.default", "end_date_type", "limited"),
				),
			},
			{
				ResourceName:      "aws_ec2_capacity_reservation.default",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSEc2CapacityReservation_tags(t *testing.T) {
	var cr ec2.CapacityReservation
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckEc2CapacityReservationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEc2CapacityReservationConfig_tags,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEc2CapacityReservationExists("aws_ec2_capacity_reservation.default", &cr),
					resource.TestCheckResourceAttr("aws_ec2_capacity_reservation.default", "tags.%", "1"),
					resource.TestCheckResourceAttr("aws_ec2_capacity_reservation.default", "tags.Name", "foo"),
				),
			},
			{
				ResourceName:      "aws_ec2_capacity_reservation.default",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckEc2CapacityReservationExists(resourceName string, cr *ec2.CapacityReservation) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).ec2conn
		resp, err := conn.DescribeCapacityReservations(&ec2.DescribeCapacityReservationsInput{
			CapacityReservationIds: []*string{aws.String(rs.Primary.ID)},
		})

		if err != nil {
			return fmt.Errorf("Error retrieving EC2 Capacity Reservations: %s", err)
		}

		if len(resp.CapacityReservations) == 0 {
			return fmt.Errorf("EC2 Capacity Reservation (%s) not found", rs.Primary.ID)
		}

		reservation := resp.CapacityReservations[0]

		if aws.StringValue(reservation.State) != ec2.CapacityReservationStateActive && aws.StringValue(reservation.State) != ec2.CapacityReservationStatePending {
			return fmt.Errorf("EC2 Capacity Reservation (%s) found in unexpected state: %s", rs.Primary.ID, aws.StringValue(reservation.State))
		}

		*cr = *reservation
		return nil
	}
}

func testAccCheckEc2CapacityReservationDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).ec2conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_ec2_capacity_reservation" {
			continue
		}

		// Try to find the resource
		resp, err := conn.DescribeCapacityReservations(&ec2.DescribeCapacityReservationsInput{
			CapacityReservationIds: []*string{aws.String(rs.Primary.ID)},
		})
		if err == nil {
			for _, r := range resp.CapacityReservations {
				if aws.StringValue(r.State) != ec2.CapacityReservationStateCancelled && aws.StringValue(r.State) != ec2.CapacityReservationStateExpired {
					return fmt.Errorf("Found uncancelled EC2 Capacity Reservation: %s", r)
				}
			}
		}

		return err
	}

	return nil

}

const testAccEc2CapacityReservationConfig = `
data "aws_availability_zones" "available" {}

resource "aws_ec2_capacity_reservation" "default" {
  instance_type     = "t2.micro"
  instance_platform = "Linux/UNIX"
  availability_zone = "${data.aws_availability_zones.available.names[0]}"
  instance_count    = 1
}
`

const testAccEc2CapacityReservationConfig_endDate = `
data "aws_availability_zones" "available" {}

resource "aws_ec2_capacity_reservation" "default" {
  instance_type     = "t2.micro"
  instance_platform = "Linux/UNIX"
  availability_zone = "${data.aws_availability_zones.available.names[0]}"
  instance_count    = 1
  end_date          = "2019-10-31T07:39:57Z"
  end_date_type     = "limited"
}
`

const testAccEc2CapacityReservationConfig_tags = `
data "aws_availability_zones" "available" {}

resource "aws_ec2_capacity_reservation" "default" {
  instance_type     = "t2.micro"
  instance_platform = "Linux/UNIX"
  availability_zone = "${data.aws_availability_zones.available.names[0]}"
  instance_count    = 1

  tags {
    Name = "Foo"
  }
}
`
