package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

// func init() {
// 	resource.AddTestSweepers("aws_capacity_reservation", &resource.Sweeper{
// 		Name: "aws_capacity_reservation",
// 		F:    testSweepCapacityReservations,
// 	})
// }

// func testSweepInstances(region string) error {
// 	client, err := sharedClientForRegion(region)
// 	if err != nil {
// 		return fmt.Errorf("error getting client: %s", err)
// 	}
// 	conn := client.(*AWSClient).ec2conn

// 	err = conn.DescribeInstancesPages(&ec2.DescribeInstancesInput{}, func(page *ec2.DescribeInstancesOutput, isLast bool) bool {
// 		if len(page.Reservations) == 0 {
// 			log.Print("[DEBUG] No EC2 Instances to sweep")
// 			return false
// 		}

// 		for _, reservation := range page.Reservations {
// 			for _, instance := range reservation.Instances {
// 				var nameTag string
// 				id := aws.StringValue(instance.InstanceId)

// 				for _, instanceTag := range instance.Tags {
// 					if aws.StringValue(instanceTag.Key) == "Name" {
// 						nameTag = aws.StringValue(instanceTag.Value)
// 						break
// 					}
// 				}

// 				if !strings.HasPrefix(nameTag, "tf-acc-test-") {
// 					log.Printf("[INFO] Skipping EC2 Instance: %s", id)
// 					continue
// 				}

// 				log.Printf("[INFO] Terminating EC2 Instance: %s", id)
// 				err := awsTerminateInstance(conn, id, 5*time.Minute)
// 				if err != nil {
// 					log.Printf("[ERROR] Error terminating EC2 Instance (%s): %s", id, err)
// 				}
// 			}
// 		}
// 		return !isLast
// 	})
// 	if err != nil {
// 		if testSweepSkipSweepError(err) {
// 			log.Printf("[WARN] Skipping EC2 Instance sweep for %s: %s", region, err)
// 			return nil
// 		}
// 		return fmt.Errorf("Error retrieving EC2 Instances: %s", err)
// 	}

// 	return nil
// }

func TestAccAWSCapacityReservation_importBasic(t *testing.T) {
	resourceName := "aws_capacity_reservation.default"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCapacityReservationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCapacityReservationConfig,
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSCapacityReservation_basic(t *testing.T) {
	var cr ec2.CapacityReservation
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCapacityReservationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCapacityReservationConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCapacityReservationExists("aws_capacity_reservation.default", &cr),
					resource.TestCheckResourceAttr("aws_capacity_reservation.default", "instance_type", "t2.micro"),
				),
			},
		},
	})
}

func TestAccAWSCapacityReservation_endDate(t *testing.T) {
	var cr ec2.CapacityReservation
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCapacityReservationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCapacityReservationConfig_endDate,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCapacityReservationExists("aws_capacity_reservation.default", &cr),
					resource.TestCheckResourceAttr("aws_capacity_reservation.default", "end_date", "2019-10-31T07:39:57Z"),
					resource.TestCheckResourceAttr("aws_capacity_reservation.default", "end_date_type", "limited"),
				),
			},
		},
	})
}

func TestAccAWSCapacityReservation_tags(t *testing.T) {
	var cr ec2.CapacityReservation
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCapacityReservationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCapacityReservationConfig_tags,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCapacityReservationExists("aws_capacity_reservation.default", &cr),
					resource.TestCheckResourceAttr("aws_capacity_reservation.default", "instance_type", "t2.micro"),
				),
			},
		},
	})
}

func testAccCheckCapacityReservationExists(resourceName string, cr *ec2.CapacityReservation) resource.TestCheckFunc {
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

func testAccCheckCapacityReservationDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).ec2conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_capacity_reservation" {
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

const testAccCapacityReservationConfig = `
resource "aws_capacity_reservation" "default" {
  instance_type     = "t2.micro"
  instance_platform = "Linux/UNIX"
  availability_zone = "us-west-2a"
  instance_count    = 1
}
`

const testAccCapacityReservationConfig_endDate = `
resource "aws_capacity_reservation" "default" {
  instance_type     = "t2.micro"
  instance_platform = "Linux/UNIX"
  availability_zone = "us-west-2a"
  instance_count    = 1
  end_date          = "2019-10-31T07:39:57Z"
  end_date_type     = "limited"
}
`

const testAccCapacityReservationConfig_tags = `
resource "aws_capacity_reservation" "default" {
  instance_type     = "t2.micro"
  instance_platform = "Linux/UNIX"
  availability_zone = "us-west-2a"
  instance_count    = 1

  tags {
    Name = "Foo"
  }
}
`
