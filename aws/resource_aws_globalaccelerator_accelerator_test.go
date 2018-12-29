package aws

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAwsGlobalAcceleratorAccelerator_basic(t *testing.T) {
	resourceName := "aws_globalaccelerator_accelerator.example"
	rInt := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckGlobalAcceleratorAcceleratorDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGlobalAcceleratorAccelerator_basic(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGlobalAcceleratorAcceleratorExists(resourceName),
					resource.TestMatchResourceAttr(resourceName, "name", regexp.MustCompile("tf-.*")),
					resource.TestCheckResourceAttr(resourceName, "enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "ip_address_type", "IPV4"),
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
	rInt := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckGlobalAcceleratorAcceleratorDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGlobalAcceleratorAccelerator_basic(rInt),
			},
			{
				Config: testAccGlobalAcceleratorAccelerator_update(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGlobalAcceleratorAcceleratorExists(resourceName),
					resource.TestMatchResourceAttr(resourceName, "name", regexp.MustCompile("tfx-.*")),
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
	rInt := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckGlobalAcceleratorAcceleratorDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGlobalAcceleratorAccelerator_attributes(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGlobalAcceleratorAcceleratorExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "attributes.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "attributes.0.flow_logs_enabled", "true"),
					resource.TestMatchResourceAttr(resourceName, "attributes.0.flow_logs_s3_bucket", regexp.MustCompile("tf-.*")),
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

func testAccGlobalAcceleratorAccelerator_basic(rInt int) string {
	return fmt.Sprintf(`
resource "aws_globalaccelerator_accelerator" "example" {
  name            = "tf-%d"
  ip_address_type = "IPV4"
  enabled         = false
}
`, rInt)
}

func testAccGlobalAcceleratorAccelerator_update(rInt int) string {
	return fmt.Sprintf(`
resource "aws_globalaccelerator_accelerator" "example" {
  name            = "tfx-%d"
  ip_address_type = "IPV4"
  enabled         = false
}
`, rInt)
}

func testAccGlobalAcceleratorAccelerator_attributes(rInt int) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "example" {
  bucket_prefix = "tf-globalaccelerator-accelerator-"
}

resource "aws_globalaccelerator_accelerator" "example" {
  name            = "tf-%d"
  ip_address_type = "IPV4"
  enabled         = false

  attributes {
    flow_logs_enabled   = true
    flow_logs_s3_bucket = "${aws_s3_bucket.example.bucket}"
    flow_logs_s3_prefix = "flow-logs/"
  }
}
`, rInt)
}
