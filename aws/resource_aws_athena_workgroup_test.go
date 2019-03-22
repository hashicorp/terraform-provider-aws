package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/athena"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAWSAthenaWorkGroup_basic(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAthenaWorkGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAthenaWorkGroupConfig(acctest.RandString(5)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAthenaWorkGroupExists("aws_athena_workgroup.foo"),
				),
			},
		},
	})
}

func TestAccAWSAthenaWorkGroup_withDescription(t *testing.T) {
	rName := acctest.RandString(5)
	rDescription := acctest.RandString(20)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAthenaWorkGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAthenaWorkGroupConfigDescription(rName, rDescription),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAthenaWorkGroupExists("aws_athena_workgroup.desc"),
					resource.TestCheckResourceAttr(
						"aws_athena_workgroup.desc", "description", rDescription),
				),
			},
		},
	})
}

func TestAccAWSAthenaWorkGroup_withDescriptionUpdate(t *testing.T) {
	rName := acctest.RandString(5)
	rDescription := acctest.RandString(20)
	rDescriptionUpdate := acctest.RandString(20)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAthenaWorkGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAthenaWorkGroupConfigDescription(rName, rDescription),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAthenaWorkGroupExists("aws_athena_workgroup.desc"),
					resource.TestCheckResourceAttr(
						"aws_athena_workgroup.desc", "description", rDescription),
				),
			},
			{
				Config: testAccAthenaWorkGroupConfigDescription(rName, rDescriptionUpdate),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAthenaWorkGroupExists("aws_athena_workgroup.desc"),
					resource.TestCheckResourceAttr(
						"aws_athena_workgroup.desc", "description", rDescriptionUpdate),
				),
			},
		},
	})
}

func TestAccAWSAthenaWorkGroup_withBytesScannedCutoffPerQuery(t *testing.T) {
	rName := acctest.RandString(5)
	rBytesScannedCutoffPerQuery := "10485760"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAthenaWorkGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAthenaWorkGroupConfigBytesScannedCutoffPerQuery(rName, rBytesScannedCutoffPerQuery),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAthenaWorkGroupExists("aws_athena_workgroup.bytes"),
					resource.TestCheckResourceAttr(
						"aws_athena_workgroup.bytes", "bytes_scanned_cutoff_per_query", rBytesScannedCutoffPerQuery),
				),
			},
		},
	})
}

func TestAccAWSAthenaWorkGroup_withBytesScannedCutoffPerQueryUpdate(t *testing.T) {
	rName := acctest.RandString(5)
	rBytesScannedCutoffPerQuery := "10485760"
	rBytesScannedCutoffPerQueryUpdate := "12582912"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAthenaWorkGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAthenaWorkGroupConfigBytesScannedCutoffPerQuery(rName, rBytesScannedCutoffPerQuery),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAthenaWorkGroupExists("aws_athena_workgroup.bytes"),
					resource.TestCheckResourceAttr(
						"aws_athena_workgroup.bytes", "bytes_scanned_cutoff_per_query", rBytesScannedCutoffPerQuery),
				),
			},
			{
				Config: testAccAthenaWorkGroupConfigBytesScannedCutoffPerQuery(rName, rBytesScannedCutoffPerQueryUpdate),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAthenaWorkGroupExists("aws_athena_workgroup.bytes"),
					resource.TestCheckResourceAttr(
						"aws_athena_workgroup.bytes", "bytes_scanned_cutoff_per_query", rBytesScannedCutoffPerQueryUpdate),
				),
			},
		},
	})
}

func TestAccAWSAthenaWorkGroup_withEnforceWorkgroupConfiguration(t *testing.T) {
	rName := acctest.RandString(5)
	rEnforce := "true"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAthenaWorkGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAthenaWorkGroupConfigEnforceWorkgroupConfiguration(rName, rEnforce),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAthenaWorkGroupExists("aws_athena_workgroup.enforce"),
					resource.TestCheckResourceAttr(
						"aws_athena_workgroup.enforce", "enforce_workgroup_configuration", rEnforce),
				),
			},
		},
	})
}

func TestAccAWSAthenaWorkGroup_withEnforceWorkgroupConfigurationUpdate(t *testing.T) {
	rName := acctest.RandString(5)
	rEnforce := "true"
	rEnforceUpdate := "false"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAthenaWorkGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAthenaWorkGroupConfigEnforceWorkgroupConfiguration(rName, rEnforce),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAthenaWorkGroupExists("aws_athena_workgroup.enforce"),
					resource.TestCheckResourceAttr(
						"aws_athena_workgroup.enforce", "enforce_workgroup_configuration", rEnforce),
				),
			},
			{
				Config: testAccAthenaWorkGroupConfigEnforceWorkgroupConfiguration(rName, rEnforceUpdate),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAthenaWorkGroupExists("aws_athena_workgroup.enforce"),
					resource.TestCheckResourceAttr(
						"aws_athena_workgroup.enforce", "enforce_workgroup_configuration", rEnforceUpdate),
				),
			},
		},
	})
}

func testAccCheckAWSAthenaWorkGroupDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).athenaconn
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_athena_workgroup" {
			continue
		}

		input := &athena.GetWorkGroupInput{
			WorkGroup: aws.String(rs.Primary.ID),
		}

		resp, err := conn.GetWorkGroup(input)
		if err != nil {
			if isAWSErr(err, athena.ErrCodeInvalidRequestException, rs.Primary.ID) {
				return nil
			}
			return err
		}
		if resp.WorkGroup != nil {
			return fmt.Errorf("Athena WorkGroup (%s) found", rs.Primary.ID)
		}
	}
	return nil
}

func testAccCheckAWSAthenaWorkGroupExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		conn := testAccProvider.Meta().(*AWSClient).athenaconn

		input := &athena.GetWorkGroupInput{
			WorkGroup: aws.String(rs.Primary.ID),
		}

		_, err := conn.GetWorkGroup(input)
		return err
	}
}

func testAccAthenaWorkGroupConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_athena_workgroup" "foo" {
  name = "tf-athena-workgroup-%s"
}
		`, rName)
}

func testAccAthenaWorkGroupConfigDescription(rName string, rDescription string) string {
	return fmt.Sprintf(`
	resource "aws_athena_workgroup" "desc" {
		name = "tf-athena-workgroup-%s"
		description = "%s"
	}
	`, rName, rDescription)
}

func testAccAthenaWorkGroupConfigBytesScannedCutoffPerQuery(rName string, rBytesScannedCutoffPerQuery string) string {
	return fmt.Sprintf(`
	resource "aws_athena_workgroup" "bytes" {
		name = "tf-athena-workgroup-%s"
		bytes_scanned_cutoff_per_query = %s
	}
	`, rName, rBytesScannedCutoffPerQuery)
}

func testAccAthenaWorkGroupConfigEnforceWorkgroupConfiguration(rName string, rEnforce string) string {
	return fmt.Sprintf(`
	resource "aws_athena_workgroup" "enforce" {
		name = "tf-athena-workgroup-%s"
		enforce_workgroup_configuration = %s
	}
	`, rName, rEnforce)
}
