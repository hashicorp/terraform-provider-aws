package aws

import (
	"fmt"
	"log"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/globalaccelerator"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func init() {
	resource.AddTestSweepers("aws_globalaccelerator_accelerator", &resource.Sweeper{
		Name: "aws_globalaccelerator_accelerator",
		F:    testSweepGlobalAcceleratorAccelerators,
	})
}

func testSweepGlobalAcceleratorAccelerators(region string) error {
	client, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("Error getting client: %s", err)
	}
	conn := client.(*AWSClient).globalacceleratorconn

	input := &globalaccelerator.ListAcceleratorsInput{}
	var sweeperErrs *multierror.Error

	for {
		output, err := conn.ListAccelerators(input)

		if testSweepSkipSweepError(err) {
			log.Printf("[WARN] Skipping Global Accelerator Accelerator sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("Error retrieving Global Accelerator Accelerators: %s", err)
		}

		for _, accelerator := range output.Accelerators {
			arn := aws.StringValue(accelerator.AcceleratorArn)

			if aws.BoolValue(accelerator.Enabled) {
				input := &globalaccelerator.UpdateAcceleratorInput{
					AcceleratorArn: accelerator.AcceleratorArn,
					Enabled:        aws.Bool(false),
				}

				log.Printf("[INFO] Disabling Global Accelerator Accelerator: %s", arn)

				_, err := conn.UpdateAccelerator(input)

				if err != nil {
					sweeperErr := fmt.Errorf("error disabling Global Accelerator Accelerator (%s): %s", arn, err)
					log.Printf("[ERROR] %s", sweeperErr)
					sweeperErrs = multierror.Append(sweeperErrs, sweeperErr)
					continue
				}

				if err := resourceAwsGlobalAcceleratorAcceleratorWaitForState(conn, arn); err != nil {
					sweeperErr := fmt.Errorf("error waiting for Global Accelerator Accelerator (%s) disable: %s", arn, err)
					log.Printf("[ERROR] %s", sweeperErr)
					sweeperErrs = multierror.Append(sweeperErrs, sweeperErr)
					continue
				}
			}

			input := &globalaccelerator.DeleteAcceleratorInput{
				AcceleratorArn: accelerator.AcceleratorArn,
			}

			log.Printf("[INFO] Deleting Global Accelerator Accelerator: %s", arn)
			_, err := conn.DeleteAccelerator(input)

			if err != nil {
				sweeperErr := fmt.Errorf("error deleting Global Accelerator Accelerator (%s): %s", arn, err)
				log.Printf("[ERROR] %s", sweeperErr)
				sweeperErrs = multierror.Append(sweeperErrs, sweeperErr)
			}
		}

		if aws.StringValue(output.NextToken) == "" {
			break
		}

		input.NextToken = output.NextToken
	}

	return sweeperErrs.ErrorOrNil()
}

func TestAccAwsGlobalAcceleratorAccelerator_basic(t *testing.T) {
	resourceName := "aws_globalaccelerator_accelerator.example"
	rName := acctest.RandomWithPrefix("tf-acc-test")
	ipRegex := regexp.MustCompile(`\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}`)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckGlobalAcceleratorAcceleratorDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGlobalAcceleratorAccelerator_basic(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGlobalAcceleratorAcceleratorExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "ip_address_type", "IPV4"),
					resource.TestCheckResourceAttr(resourceName, "attributes.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "attributes.0.flow_logs_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "attributes.0.flow_logs_s3_bucket", ""),
					resource.TestCheckResourceAttr(resourceName, "attributes.0.flow_logs_s3_prefix", ""),
					resource.TestCheckResourceAttr(resourceName, "ip_sets.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "ip_sets.0.ip_addresses.#", "2"),
					resource.TestMatchResourceAttr(resourceName, "ip_sets.0.ip_addresses.0", ipRegex),
					resource.TestMatchResourceAttr(resourceName, "ip_sets.0.ip_addresses.1", ipRegex),
					resource.TestCheckResourceAttr(resourceName, "ip_sets.0.ip_family", "IPv4"),
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

func TestAccAwsGlobalAcceleratorAccelerator_update(t *testing.T) {
	resourceName := "aws_globalaccelerator_accelerator.example"
	rName := acctest.RandomWithPrefix("tf-acc-test")
	newName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckGlobalAcceleratorAcceleratorDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGlobalAcceleratorAccelerator_basic(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGlobalAcceleratorAcceleratorExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "enabled", "true"),
				),
			},
			{
				Config: testAccGlobalAcceleratorAccelerator_basic(newName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGlobalAcceleratorAcceleratorExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", newName),
					resource.TestCheckResourceAttr(resourceName, "enabled", "false"),
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

func TestAccAwsGlobalAcceleratorAccelerator_attributes(t *testing.T) {
	resourceName := "aws_globalaccelerator_accelerator.example"
	s3BucketResourceName := "aws_s3_bucket.example"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckGlobalAcceleratorAcceleratorDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGlobalAcceleratorAccelerator_attributes(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGlobalAcceleratorAcceleratorExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "attributes.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "attributes.0.flow_logs_enabled", "true"),
					resource.TestCheckResourceAttrPair(resourceName, "attributes.0.flow_logs_s3_bucket", s3BucketResourceName, "bucket"),
					resource.TestCheckResourceAttr(resourceName, "attributes.0.flow_logs_s3_prefix", "flow-logs/"),
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

func testAccCheckGlobalAcceleratorAcceleratorExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).globalacceleratorconn

		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		accelerator, err := resourceAwsGlobalAcceleratorAcceleratorRetrieve(conn, rs.Primary.ID)
		if err != nil {
			return err
		}

		if accelerator == nil {
			return fmt.Errorf("Global Accelerator accelerator not found")
		}

		return nil
	}
}

func testAccCheckGlobalAcceleratorAcceleratorDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).globalacceleratorconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_globalaccelerator_accelerator" {
			continue
		}

		accelerator, err := resourceAwsGlobalAcceleratorAcceleratorRetrieve(conn, rs.Primary.ID)
		if err != nil {
			return err
		}

		if accelerator != nil {
			return fmt.Errorf("Global Accelerator accelerator still exists")
		}
	}
	return nil
}

func testAccGlobalAcceleratorAccelerator_basic(rName string, enabled bool) string {
	return fmt.Sprintf(`
resource "aws_globalaccelerator_accelerator" "example" {
  name            = "%s"
  ip_address_type = "IPV4"
  enabled         = %t
}
`, rName, enabled)
}

func testAccGlobalAcceleratorAccelerator_attributes(rName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "example" {
  bucket_prefix = "tf-globalaccelerator-accelerator-"
}

resource "aws_globalaccelerator_accelerator" "example" {
  name            = "%s"
  ip_address_type = "IPV4"
  enabled         = false

  attributes {
    flow_logs_enabled   = true
    flow_logs_s3_bucket = "${aws_s3_bucket.example.bucket}"
    flow_logs_s3_prefix = "flow-logs/"
  }
}
`, rName)
}
