package ssm_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfssm "github.com/hashicorp/terraform-provider-aws/internal/service/ssm"
)

func TestAccSSMMaintenanceWindowTarget_basic(t *testing.T) {
	var maint ssm.MaintenanceWindowTarget
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ssm_maintenance_window_target.test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ssm.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckMaintenanceWindowTargetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccMaintenanceWindowTargetConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMaintenanceWindowTargetExists(resourceName, &maint),
					resource.TestCheckResourceAttr(resourceName, "targets.0.key", "tag:Name"),
					resource.TestCheckResourceAttr(resourceName, "targets.0.values.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "targets.0.values.0", "acceptance_test"),
					resource.TestCheckResourceAttr(resourceName, "targets.1.key", "tag:Name2"),
					resource.TestCheckResourceAttr(resourceName, "targets.1.values.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "targets.1.values.0", "acceptance_test"),
					resource.TestCheckResourceAttr(resourceName, "targets.1.values.1", "acceptance_test2"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "description", "This resource is for test purpose only"),
					resource.TestCheckResourceAttr(resourceName, "resource_type", ssm.MaintenanceWindowResourceTypeInstance),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateIdFunc: testAccMaintenanceWindowTargetImportStateIdFunc(resourceName),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccSSMMaintenanceWindowTarget_noNameOrDescription(t *testing.T) {
	var maint ssm.MaintenanceWindowTarget
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ssm_maintenance_window_target.test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ssm.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckMaintenanceWindowTargetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccMaintenanceWindowTargetConfig_noNameOrDescription(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMaintenanceWindowTargetExists(resourceName, &maint),
					resource.TestCheckResourceAttr(resourceName, "targets.0.key", "tag:Name"),
					resource.TestCheckResourceAttr(resourceName, "targets.0.values.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "targets.0.values.0", "acceptance_test"),
					resource.TestCheckResourceAttr(resourceName, "targets.1.key", "tag:Name2"),
					resource.TestCheckResourceAttr(resourceName, "targets.1.values.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "targets.1.values.0", "acceptance_test"),
					resource.TestCheckResourceAttr(resourceName, "targets.1.values.1", "acceptance_test2"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateIdFunc: testAccMaintenanceWindowTargetImportStateIdFunc(resourceName),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccSSMMaintenanceWindowTarget_validation(t *testing.T) {
	name := sdkacctest.RandString(10)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ssm.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckMaintenanceWindowTargetDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccMaintenanceWindowTargetConfig_basic2(name, "Bäd Name!@#$%^", "good description"),
				ExpectError: regexp.MustCompile(`Only alphanumeric characters, hyphens, dots & underscores allowed`),
			},
			{
				Config:      testAccMaintenanceWindowTargetConfig_basic2(name, "goodname", "bd"),
				ExpectError: regexp.MustCompile(`expected length of [\w]+ to be in the range \(3 - 128\), got [\w]+`),
			},
			{
				Config:      testAccMaintenanceWindowTargetConfig_basic2(name, "goodname", "This description is tooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooo long"),
				ExpectError: regexp.MustCompile(`expected length of [\w]+ to be in the range \(3 - 128\), got [\w]+`),
			},
		},
	})
}

func TestAccSSMMaintenanceWindowTarget_update(t *testing.T) {
	var maint ssm.MaintenanceWindowTarget
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ssm_maintenance_window_target.test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ssm.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckMaintenanceWindowTargetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccMaintenanceWindowTargetConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMaintenanceWindowTargetExists(resourceName, &maint),
					resource.TestCheckResourceAttr(resourceName, "targets.0.key", "tag:Name"),
					resource.TestCheckResourceAttr(resourceName, "targets.0.values.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "targets.0.values.0", "acceptance_test"),
					resource.TestCheckResourceAttr(resourceName, "targets.1.key", "tag:Name2"),
					resource.TestCheckResourceAttr(resourceName, "targets.1.values.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "targets.1.values.0", "acceptance_test"),
					resource.TestCheckResourceAttr(resourceName, "targets.1.values.1", "acceptance_test2"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "description", "This resource is for test purpose only"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateIdFunc: testAccMaintenanceWindowTargetImportStateIdFunc(resourceName),
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccMaintenanceWindowTargetConfig_basicUpdated(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMaintenanceWindowTargetExists(resourceName, &maint),
					resource.TestCheckResourceAttr(resourceName, "owner_information", "something"),
					resource.TestCheckResourceAttr(resourceName, "targets.0.key", "tag:Name"),
					resource.TestCheckResourceAttr(resourceName, "targets.0.values.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "targets.0.values.0", "acceptance_test"),
					resource.TestCheckResourceAttr(resourceName, "targets.1.key", "tag:Updated"),
					resource.TestCheckResourceAttr(resourceName, "targets.1.values.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "targets.1.values.0", "new-value"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "description", "This resource is for test purpose only - updated"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateIdFunc: testAccMaintenanceWindowTargetImportStateIdFunc(resourceName),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccSSMMaintenanceWindowTarget_resourceGroup(t *testing.T) {
	var maint ssm.MaintenanceWindowTarget
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ssm_maintenance_window_target.test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ssm.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckMaintenanceWindowTargetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccMaintenanceWindowTargetConfig_basicResourceGroup(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMaintenanceWindowTargetExists(resourceName, &maint),
					resource.TestCheckResourceAttr(resourceName, "targets.0.key", "resource-groups:ResourceTypeFilters"),
					resource.TestCheckResourceAttr(resourceName, "targets.0.values.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "targets.0.values.0", "AWS::EC2::Instance"),
					resource.TestCheckResourceAttr(resourceName, "targets.1.key", "resource-groups:Name"),
					resource.TestCheckResourceAttr(resourceName, "targets.1.values.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "targets.1.values.0", "resource-group-name"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "description", "This resource is for test purpose only"),
					resource.TestCheckResourceAttr(resourceName, "resource_type", ssm.MaintenanceWindowResourceTypeResourceGroup),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateIdFunc: testAccMaintenanceWindowTargetImportStateIdFunc(resourceName),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccSSMMaintenanceWindowTarget_disappears(t *testing.T) {
	var maint ssm.MaintenanceWindowTarget
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ssm_maintenance_window_target.test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ssm.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckMaintenanceWindowTargetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccMaintenanceWindowTargetConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMaintenanceWindowTargetExists(resourceName, &maint),
					acctest.CheckResourceDisappears(acctest.Provider, tfssm.ResourceMaintenanceWindowTarget(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccSSMMaintenanceWindowTarget_Disappears_window(t *testing.T) {
	var maint ssm.MaintenanceWindowTarget
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ssm_maintenance_window_target.test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ssm.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckMaintenanceWindowTargetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccMaintenanceWindowTargetConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMaintenanceWindowTargetExists(resourceName, &maint),
					acctest.CheckResourceDisappears(acctest.Provider, tfssm.ResourceMaintenanceWindow(), "aws_ssm_maintenance_window.test"),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckMaintenanceWindowTargetExists(n string, mWindTarget *ssm.MaintenanceWindowTarget) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No SSM Maintenance Window Target Window ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).SSMConn

		resp, err := conn.DescribeMaintenanceWindowTargets(&ssm.DescribeMaintenanceWindowTargetsInput{
			WindowId: aws.String(rs.Primary.Attributes["window_id"]),
			Filters: []*ssm.MaintenanceWindowFilter{
				{
					Key:    aws.String("WindowTargetId"),
					Values: []*string{aws.String(rs.Primary.ID)},
				},
			},
		})
		if err != nil {
			return err
		}

		for _, i := range resp.Targets {
			if aws.StringValue(i.WindowTargetId) == rs.Primary.ID {
				*mWindTarget = *resp.Targets[0]
				return nil
			}
		}

		return fmt.Errorf("No AWS SSM Maintenance window target found")
	}
}

func testAccCheckMaintenanceWindowTargetDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).SSMConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_ssm_maintenance_window_target" {
			continue
		}

		out, err := conn.DescribeMaintenanceWindowTargets(&ssm.DescribeMaintenanceWindowTargetsInput{
			WindowId: aws.String(rs.Primary.Attributes["window_id"]),
			Filters: []*ssm.MaintenanceWindowFilter{
				{
					Key:    aws.String("WindowTargetId"),
					Values: []*string{aws.String(rs.Primary.ID)},
				},
			},
		})

		if err != nil {
			// Verify the error is what we want
			if tfawserr.ErrCodeEquals(err, ssm.ErrCodeDoesNotExistException) {
				continue
			}
			return err
		}

		if len(out.Targets) > 0 {
			return fmt.Errorf("Expected AWS SSM Maintenance Target to be gone, but was still found")
		}

		return nil
	}

	return nil
}

func testAccMaintenanceWindowTargetConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_ssm_maintenance_window" "test" {
  name     = %[1]q
  schedule = "cron(0 16 ? * TUE *)"
  duration = 3
  cutoff   = 1
}

resource "aws_ssm_maintenance_window_target" "test" {
  name          = %[1]q
  description   = "This resource is for test purpose only"
  window_id     = aws_ssm_maintenance_window.test.id
  resource_type = "INSTANCE"

  targets {
    key    = "tag:Name"
    values = ["acceptance_test"]
  }

  targets {
    key    = "tag:Name2"
    values = ["acceptance_test", "acceptance_test2"]
  }
}
`, rName)
}

func testAccMaintenanceWindowTargetConfig_basicResourceGroup(rName string) string {
	return fmt.Sprintf(`
resource "aws_ssm_maintenance_window" "test" {
  name     = %[1]q
  schedule = "cron(0 16 ? * TUE *)"
  duration = 3
  cutoff   = 1
}

resource "aws_ssm_maintenance_window_target" "test" {
  name          = %[1]q
  description   = "This resource is for test purpose only"
  window_id     = aws_ssm_maintenance_window.test.id
  resource_type = "RESOURCE_GROUP"

  targets {
    key    = "resource-groups:ResourceTypeFilters"
    values = ["AWS::EC2::Instance"]
  }

  targets {
    key    = "resource-groups:Name"
    values = ["resource-group-name"]
  }
}
`, rName)
}

func testAccMaintenanceWindowTargetConfig_noNameOrDescription(rName string) string {
	return fmt.Sprintf(`
resource "aws_ssm_maintenance_window" "test" {
  name     = %[1]q
  schedule = "cron(0 16 ? * TUE *)"
  duration = 3
  cutoff   = 1
}

resource "aws_ssm_maintenance_window_target" "test" {
  window_id     = aws_ssm_maintenance_window.test.id
  resource_type = "INSTANCE"

  targets {
    key    = "tag:Name"
    values = ["acceptance_test"]
  }

  targets {
    key    = "tag:Name2"
    values = ["acceptance_test", "acceptance_test2"]
  }
}
`, rName)
}

func testAccMaintenanceWindowTargetConfig_basicUpdated(rName string) string {
	return fmt.Sprintf(`
resource "aws_ssm_maintenance_window" "test" {
  name     = %[1]q
  schedule = "cron(0 16 ? * TUE *)"
  duration = 3
  cutoff   = 1
}

resource "aws_ssm_maintenance_window_target" "test" {
  name              = %[1]q
  description       = "This resource is for test purpose only - updated"
  window_id         = aws_ssm_maintenance_window.test.id
  resource_type     = "INSTANCE"
  owner_information = "something"

  targets {
    key    = "tag:Name"
    values = ["acceptance_test"]
  }

  targets {
    key    = "tag:Updated"
    values = ["new-value"]
  }
}
`, rName)
}

func testAccMaintenanceWindowTargetConfig_basic2(rName, tName, tDesc string) string {
	return fmt.Sprintf(`
resource "aws_ssm_maintenance_window" "test" {
  name     = %[1]q
  schedule = "cron(0 16 ? * TUE *)"
  duration = 3
  cutoff   = 1
}

resource "aws_ssm_maintenance_window_target" "test" {
  name              = %[2]q
  description       = %[3]q
  window_id         = aws_ssm_maintenance_window.test.id
  resource_type     = "INSTANCE"
  owner_information = "something"

  targets {
    key    = "tag:Name"
    values = ["acceptance_test"]
  }

  targets {
    key    = "tag:Updated"
    values = ["new-value"]
  }
}
`, rName, tName, tDesc)
}

func testAccMaintenanceWindowTargetImportStateIdFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("Not found: %s", resourceName)
		}

		return fmt.Sprintf("%s/%s", rs.Primary.Attributes["window_id"], rs.Primary.ID), nil
	}
}
