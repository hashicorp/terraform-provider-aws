package aws

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAWSSSMMaintenanceWindowTarget_basic(t *testing.T) {
	name := acctest.RandString(10)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSSMMaintenanceWindowTargetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSSMMaintenanceWindowTargetBasicConfig(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSSMMaintenanceWindowTargetExists("aws_ssm_maintenance_window_target.target"),
					resource.TestCheckResourceAttr("aws_ssm_maintenance_window_target.target", "targets.0.key", "tag:Name"),
					resource.TestCheckResourceAttr("aws_ssm_maintenance_window_target.target", "targets.0.values.#", "1"),
					resource.TestCheckResourceAttr("aws_ssm_maintenance_window_target.target", "targets.0.values.0", "acceptance_test"),
					resource.TestCheckResourceAttr("aws_ssm_maintenance_window_target.target", "targets.1.key", "tag:Name2"),
					resource.TestCheckResourceAttr("aws_ssm_maintenance_window_target.target", "targets.1.values.#", "2"),
					resource.TestCheckResourceAttr("aws_ssm_maintenance_window_target.target", "targets.1.values.0", "acceptance_test"),
					resource.TestCheckResourceAttr("aws_ssm_maintenance_window_target.target", "targets.1.values.1", "acceptance_test2"),
					resource.TestCheckResourceAttr("aws_ssm_maintenance_window_target.target", "name", "TestMaintenanceWindowTarget"),
					resource.TestCheckResourceAttr("aws_ssm_maintenance_window_target.target", "description", "This resource is for test purpose only"),
				),
			},
			{
				ResourceName:      "aws_ssm_maintenance_window.foo",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSSSMMaintenanceWindowTarget_noNameOrDescription(t *testing.T) {
	name := acctest.RandString(10)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSSMMaintenanceWindowTargetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSSMMaintenanceWindowTargetNoNameOrDescriptionConfig(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSSMMaintenanceWindowTargetExists("aws_ssm_maintenance_window_target.target"),
					resource.TestCheckResourceAttr("aws_ssm_maintenance_window_target.target", "targets.0.key", "tag:Name"),
					resource.TestCheckResourceAttr("aws_ssm_maintenance_window_target.target", "targets.0.values.#", "1"),
					resource.TestCheckResourceAttr("aws_ssm_maintenance_window_target.target", "targets.0.values.0", "acceptance_test"),
					resource.TestCheckResourceAttr("aws_ssm_maintenance_window_target.target", "targets.1.key", "tag:Name2"),
					resource.TestCheckResourceAttr("aws_ssm_maintenance_window_target.target", "targets.1.values.#", "2"),
					resource.TestCheckResourceAttr("aws_ssm_maintenance_window_target.target", "targets.1.values.0", "acceptance_test"),
					resource.TestCheckResourceAttr("aws_ssm_maintenance_window_target.target", "targets.1.values.1", "acceptance_test2"),
				),
			},
			{
				ResourceName:      "aws_ssm_maintenance_window.foo",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSSSMMaintenanceWindowTarget_validation(t *testing.T) {
	name := acctest.RandString(10)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSSMMaintenanceWindowTargetDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccAWSSSMMaintenanceWindowTargetBasicConfigWithTarget(name, "Bäd Name!@#$%^", "good description"),
				ExpectError: regexp.MustCompile(`Only alphanumeric characters, hyphens, dots & underscores allowed`),
			},
			{
				Config:      testAccAWSSSMMaintenanceWindowTargetBasicConfigWithTarget(name, "goodname", "bd"),
				ExpectError: regexp.MustCompile(`expected length of [\w]+ to be in the range \(3 - 128\), got [\w]+`),
			},
			{
				Config:      testAccAWSSSMMaintenanceWindowTargetBasicConfigWithTarget(name, "goodname", "This description is tooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooo long"),
				ExpectError: regexp.MustCompile(`expected length of [\w]+ to be in the range \(3 - 128\), got [\w]+`),
			},
		},
	})
}

func TestAccAWSSSMMaintenanceWindowTarget_update(t *testing.T) {
	name := acctest.RandString(10)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSSMMaintenanceWindowTargetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSSMMaintenanceWindowTargetBasicConfig(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSSMMaintenanceWindowTargetExists("aws_ssm_maintenance_window_target.target"),
					resource.TestCheckResourceAttr("aws_ssm_maintenance_window_target.target", "targets.0.key", "tag:Name"),
					resource.TestCheckResourceAttr("aws_ssm_maintenance_window_target.target", "targets.0.values.#", "1"),
					resource.TestCheckResourceAttr("aws_ssm_maintenance_window_target.target", "targets.0.values.0", "acceptance_test"),
					resource.TestCheckResourceAttr("aws_ssm_maintenance_window_target.target", "targets.1.key", "tag:Name2"),
					resource.TestCheckResourceAttr("aws_ssm_maintenance_window_target.target", "targets.1.values.#", "2"),
					resource.TestCheckResourceAttr("aws_ssm_maintenance_window_target.target", "targets.1.values.0", "acceptance_test"),
					resource.TestCheckResourceAttr("aws_ssm_maintenance_window_target.target", "targets.1.values.1", "acceptance_test2"),
					resource.TestCheckResourceAttr("aws_ssm_maintenance_window_target.target", "name", "TestMaintenanceWindowTarget"),
					resource.TestCheckResourceAttr("aws_ssm_maintenance_window_target.target", "description", "This resource is for test purpose only"),
				),
			},
			{
				Config: testAccAWSSSMMaintenanceWindowTargetBasicConfigUpdated(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSSMMaintenanceWindowTargetExists("aws_ssm_maintenance_window_target.target"),
					resource.TestCheckResourceAttr("aws_ssm_maintenance_window_target.target", "owner_information", "something"),
					resource.TestCheckResourceAttr("aws_ssm_maintenance_window_target.target", "targets.0.key", "tag:Name"),
					resource.TestCheckResourceAttr("aws_ssm_maintenance_window_target.target", "targets.0.values.#", "1"),
					resource.TestCheckResourceAttr("aws_ssm_maintenance_window_target.target", "targets.0.values.0", "acceptance_test"),
					resource.TestCheckResourceAttr("aws_ssm_maintenance_window_target.target", "targets.1.key", "tag:Updated"),
					resource.TestCheckResourceAttr("aws_ssm_maintenance_window_target.target", "targets.1.values.#", "1"),
					resource.TestCheckResourceAttr("aws_ssm_maintenance_window_target.target", "targets.1.values.0", "new-value"),
					resource.TestCheckResourceAttr("aws_ssm_maintenance_window_target.target", "name", "TestMaintenanceWindowTarget"),
					resource.TestCheckResourceAttr("aws_ssm_maintenance_window_target.target", "description", "This resource is for test purpose only - updated"),
				),
			},
		},
	})
}

func testAccCheckAWSSSMMaintenanceWindowTargetExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No SSM Maintenance Window Target Window ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).ssmconn

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
			if *i.WindowTargetId == rs.Primary.ID {
				return nil
			}
		}

		return fmt.Errorf("No AWS SSM Maintenance window target found")
	}
}

func testAccCheckAWSSSMMaintenanceWindowTargetDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).ssmconn

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
			if ae, ok := err.(awserr.Error); ok && ae.Code() == "DoesNotExistException" {
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

func testAccAWSSSMMaintenanceWindowTargetBasicConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_ssm_maintenance_window" "foo" {
  name     = "maintenance-window-%s"
  schedule = "cron(0 16 ? * TUE *)"
  duration = 3
  cutoff   = 1
}

resource "aws_ssm_maintenance_window_target" "target" {
  name          = "TestMaintenanceWindowTarget"
  description   = "This resource is for test purpose only"
  window_id     = "${aws_ssm_maintenance_window.foo.id}"
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

func testAccAWSSSMMaintenanceWindowTargetNoNameOrDescriptionConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_ssm_maintenance_window" "foo" {
  name     = "maintenance-window-%s"
  schedule = "cron(0 16 ? * TUE *)"
  duration = 3
  cutoff   = 1
}

resource "aws_ssm_maintenance_window_target" "target" {
  window_id     = "${aws_ssm_maintenance_window.foo.id}"
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

func testAccAWSSSMMaintenanceWindowTargetBasicConfigUpdated(rName string) string {
	return fmt.Sprintf(`
resource "aws_ssm_maintenance_window" "foo" {
  name     = "maintenance-window-%s"
  schedule = "cron(0 16 ? * TUE *)"
  duration = 3
  cutoff   = 1
}

resource "aws_ssm_maintenance_window_target" "target" {
  name              = "TestMaintenanceWindowTarget"
  description       = "This resource is for test purpose only - updated"
  window_id         = "${aws_ssm_maintenance_window.foo.id}"
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

func testAccAWSSSMMaintenanceWindowTargetBasicConfigWithTarget(rName, tName, tDesc string) string {
	return fmt.Sprintf(`
resource "aws_ssm_maintenance_window" "foo" {
  name     = "maintenance-window-%s"
  schedule = "cron(0 16 ? * TUE *)"
  duration = 3
  cutoff   = 1
}

resource "aws_ssm_maintenance_window_target" "target" {
  name              = "%s"
  description       = "%s"
  window_id         = "${aws_ssm_maintenance_window.foo.id}"
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
