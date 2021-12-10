package ecr_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/ecr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func TestAccECRScanningConfiguration_serial(t *testing.T) {
	testFuncs := map[string]func(t *testing.T){
		"basic": testAccRegistryScanningConfiguration_basic,
		// "disappears": testAccRegistryPolicy_disappears,
	}

	for name, testFunc := range testFuncs {
		testFunc := testFunc

		t.Run(name, func(t *testing.T) {
			testFunc(t)
		})
	}
}

func testAccRegistryScanningConfiguration_basic(t *testing.T) {
	var v ecr.GetRegistryScanningConfigurationOutput
	resourceName := "aws_ecr_registry_scanning_configuration.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ecr.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccRegistryScanningConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRegistryScanningConfiguration(),
				Check: resource.ComposeTestCheckFunc(
					testAccRegistryScanningConfigurationExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "scan_type", "ENHANCED"),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccRegistryPolicyUpdated(),
				Check: resource.ComposeTestCheckFunc(
					testAccRegistryScanningConfigurationExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "scan_type", "BASIC"),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "0"),
				),
			},
		},
	})
}

// func testAccRegistryScanningConfiguration_disappears(t *testing.T) {
// 	var v ecr.GetRegistryPolicyOutput
// 	resourceName := "aws_ecr_registry_policy.test"

// 	resource.Test(t, resource.TestCase{
// 		PreCheck:     func() { acctest.PreCheck(t) },
// 		ErrorCheck:   acctest.ErrorCheck(t, ecr.EndpointsID),
// 		Providers:    acctest.Providers,
// 		CheckDestroy: testAccCheckRegistryPolicyDestroy,
// 		Steps: []resource.TestStep{
// 			{
// 				Config: testAccRegistryPolicy(),
// 				Check: resource.ComposeTestCheckFunc(
// 					testAccCheckRegistryPolicyExists(resourceName, &v),
// 					acctest.CheckResourceDisappears(acctest.Provider, tfecr.ResourceRegistryPolicy(), resourceName),
// 				),
// 				ExpectNonEmptyPlan: true,
// 			},
// 		},
// 	})
// }

func testAccRegistryScanningConfigurationDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).ECRConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_ecr_registry_policy" {
			continue
		}

		_, err := conn.GetRegistryScanningConfiguration(&ecr.GetRegistryScanningConfigurationInput{})
		if err != nil {
			return err
		}
	}

	return nil
}

func testAccRegistryScanningConfigurationExists(name string, res *ecr.GetRegistryScanningConfigurationOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ECR registry policy ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).ECRConn

		output, err := conn.GetRegistryScanningConfiguration(&ecr.GetRegistryScanningConfigurationInput{})
		if err != nil {
			return err
		}

		*res = *output

		return nil
	}
}

func testAccRegistryScanningConfiguration() string {
	return `
resource "aws_ecr_registry_scanning_configuration" "test" {
  scan_type = "ENHANCED"
  rule {
    scan_frequency = "CONTINUOUS_SCAN"
    repository_filter {
      filter      = "example"
      filter_type = "WILDCARD"
    }
  }
}
`
}

func testAccRegistryScanningConfigurationUpdated() string {
	return `
resource "aws_ecr_registry_scanning_configuration" "test" {
  scan_type = "BASIC"
}
`
}
