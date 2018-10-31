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
		return fmt.Errorf("Error retrieving capacity reservations: %s", err)
	}

	if len(resp.CapacityReservations) == 0 {
		log.Print("[DEBUG] No capacity reservations to sweep")
		return nil
	}

	for _, r := range resp.CapacityReservations {
		if *r.State != "cancelled" {
			id := aws.StringValue(r.CapacityReservationId)

			log.Printf("[INFO] Cancelling capacity reservation EC2 Instance: %s", id)

			opts := &ec2.CancelCapacityReservationInput{
				CapacityReservationId: aws.String(id),
			}

			_, err := conn.CancelCapacityReservation(opts)

			if err != nil {
				log.Printf("[ERROR] Error cancelling capacity reservation (%s): %s", id, err)
			}
		}
	}

	return nil
}

func TestAccAWSEc2CapacityReservation_importBasic(t *testing.T) {
	resourceName := "aws_ec2_capacity_reservation.default"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckEc2CapacityReservationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEc2CapacityReservationConfig,
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
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
					resource.TestCheckResourceAttr("aws_ec2_capacity_reservation.default", "instance_type", "t2.micro"),
				),
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
			return fmt.Errorf("DescribeCapacityReservations error: %v", err)
		}

		if len(resp.CapacityReservations) > 0 {
			*cr = *resp.CapacityReservations[0]
			return nil
		}

		return fmt.Errorf("Capacity Reservation not found")
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
				if *r.State != "cancelled" {
					return fmt.Errorf("Found uncancelled Capacity Reservation: %s", r)
				}
			}
		}

		return err
	}

	return nil

}

const testAccEc2CapacityReservationConfig = `
resource "aws_ec2_capacity_reservation" "default" {
  instance_type     = "t2.micro"
  instance_platform = "Linux/UNIX"
  availability_zone = "us-west-2a"
  instance_count    = 1
}
`

const testAccEc2CapacityReservationConfig_endDate = `
resource "aws_ec2_capacity_reservation" "default" {
  instance_type     = "t2.micro"
  instance_platform = "Linux/UNIX"
  availability_zone = "us-west-2a"
  instance_count    = 1
  end_date          = "2019-10-31T07:39:57Z"
  end_date_type     = "limited"
}
`

const testAccEc2CapacityReservationConfig_tags = `
resource "aws_ec2_capacity_reservation" "default" {
  instance_type     = "t2.micro"
  instance_platform = "Linux/UNIX"
  availability_zone = "us-west-2a"
  instance_count    = 1

  tags {
    Name = "Foo"
  }
}
`
