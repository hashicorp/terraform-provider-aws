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
	var workgroup1 athena.WorkGroup
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_athena_workgroup.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAthenaWorkGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAthenaWorkGroupConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAthenaWorkGroupExists(resourceName, &workgroup1),
					resource.TestCheckResourceAttr(resourceName, "state", athena.WorkGroupStateEnabled),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
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

func TestAccAWSAthenaWorkGroup_disappears(t *testing.T) {
	var workgroup1 athena.WorkGroup
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_athena_workgroup.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAthenaWorkGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAthenaWorkGroupConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAthenaWorkGroupExists(resourceName, &workgroup1),
					testAccCheckAWSAthenaWorkGroupDisappears(&workgroup1),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSAthenaWorkGroup_Description(t *testing.T) {
	var workgroup1, workgroup2 athena.WorkGroup
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_athena_workgroup.test"
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
					testAccCheckAWSAthenaWorkGroupExists(resourceName, &workgroup1),
					resource.TestCheckResourceAttr(resourceName, "description", rDescription),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAthenaWorkGroupConfigDescription(rName, rDescriptionUpdate),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAthenaWorkGroupExists(resourceName, &workgroup2),
					resource.TestCheckResourceAttr(resourceName, "description", rDescriptionUpdate),
				),
			},
		},
	})
}

func TestAccAWSAthenaWorkGroup_BytesScannedCutoffPerQuery(t *testing.T) {
	var workgroup1, workgroup2 athena.WorkGroup
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_athena_workgroup.test"
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
					testAccCheckAWSAthenaWorkGroupExists(resourceName, &workgroup1),
					resource.TestCheckResourceAttr(resourceName, "bytes_scanned_cutoff_per_query", rBytesScannedCutoffPerQuery),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAthenaWorkGroupConfigBytesScannedCutoffPerQuery(rName, rBytesScannedCutoffPerQueryUpdate),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAthenaWorkGroupExists(resourceName, &workgroup2),
					resource.TestCheckResourceAttr(resourceName, "bytes_scanned_cutoff_per_query", rBytesScannedCutoffPerQueryUpdate),
				),
			},
		},
	})
}

func TestAccAWSAthenaWorkGroup_EnforceWorkgroupConfiguration(t *testing.T) {
	var workgroup1, workgroup2 athena.WorkGroup
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_athena_workgroup.test"
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
					testAccCheckAWSAthenaWorkGroupExists(resourceName, &workgroup1),
					resource.TestCheckResourceAttr(resourceName, "enforce_workgroup_configuration", rEnforce),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAthenaWorkGroupConfigEnforceWorkgroupConfiguration(rName, rEnforceUpdate),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAthenaWorkGroupExists(resourceName, &workgroup2),
					resource.TestCheckResourceAttr(resourceName, "enforce_workgroup_configuration", rEnforceUpdate),
				),
			},
		},
	})
}

func TestAccAWSAthenaWorkGroup_PublishCloudWatchMetricsEnabled(t *testing.T) {
	var workgroup1, workgroup2 athena.WorkGroup
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_athena_workgroup.test"
	rEnabled := "true"
	rEnabledUpdate := "false"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAthenaWorkGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAthenaWorkGroupConfigPublishCloudWatchMetricsEnabled(rName, rEnabled),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAthenaWorkGroupExists(resourceName, &workgroup1),
					resource.TestCheckResourceAttr(resourceName, "publish_cloudwatch_metrics_enabled", rEnabled),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAthenaWorkGroupConfigPublishCloudWatchMetricsEnabled(rName, rEnabledUpdate),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAthenaWorkGroupExists(resourceName, &workgroup2),
					resource.TestCheckResourceAttr(resourceName, "publish_cloudwatch_metrics_enabled", rEnabledUpdate),
				),
			},
		},
	})
}

func TestAccAWSAthenaWorkGroup_OutputLocation(t *testing.T) {
	var workgroup1, workgroup2 athena.WorkGroup
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_athena_workgroup.test"
	rOutputLocation1 := fmt.Sprintf("%s-1", rName)
	rOutputLocation2 := fmt.Sprintf("%s-2", rName)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAthenaWorkGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAthenaWorkGroupConfigOutputLocation(rName, rOutputLocation1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAthenaWorkGroupExists(resourceName, &workgroup1),
					resource.TestCheckResourceAttr(resourceName, "output_location", "s3://"+rOutputLocation1+"/test/output"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAthenaWorkGroupConfigOutputLocation(rName, rOutputLocation2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAthenaWorkGroupExists(resourceName, &workgroup2),
					resource.TestCheckResourceAttr(resourceName, "output_location", "s3://"+rOutputLocation2+"/test/output"),
				),
			},
		},
	})
}

func TestAccAWSAthenaWorkGroup_SseS3Encryption(t *testing.T) {
	var workgroup1 athena.WorkGroup
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_athena_workgroup.test"
	rEncryption := athena.EncryptionOptionSseS3

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAthenaWorkGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAthenaWorkGroupConfigEncryptionS3(rName, rEncryption),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAthenaWorkGroupExists(resourceName, &workgroup1),
					resource.TestCheckResourceAttr(resourceName, "encryption_option", rEncryption),
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

func TestAccAWSAthenaWorkGroup_KmsEncryption(t *testing.T) {
	var workgroup1, workgroup2 athena.WorkGroup
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_athena_workgroup.test"
	rEncryption := athena.EncryptionOptionSseKms
	rEncryption2 := athena.EncryptionOptionCseKms

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAthenaWorkGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAthenaWorkGroupConfigEncryptionKms(rName, rEncryption),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAthenaWorkGroupExists(resourceName, &workgroup1),
					resource.TestCheckResourceAttr(resourceName, "encryption_option", rEncryption),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAthenaWorkGroupConfigEncryptionKms(rName, rEncryption2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAthenaWorkGroupExists(resourceName, &workgroup2),
					resource.TestCheckResourceAttr(resourceName, "encryption_option", rEncryption2),
				),
			},
		},
	})
}

func TestAccAWSAthenaWorkGroup_State(t *testing.T) {
	var workgroup1, workgroup2, workgroup3 athena.WorkGroup
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_athena_workgroup.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAthenaWorkGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAthenaWorkGroupConfigState(rName, athena.WorkGroupStateDisabled),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAthenaWorkGroupExists(resourceName, &workgroup1),
					resource.TestCheckResourceAttr(resourceName, "state", athena.WorkGroupStateDisabled),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAthenaWorkGroupConfigState(rName, athena.WorkGroupStateEnabled),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAthenaWorkGroupExists(resourceName, &workgroup2),
					resource.TestCheckResourceAttr(resourceName, "state", athena.WorkGroupStateEnabled),
				),
			},
			{
				Config: testAccAthenaWorkGroupConfigState(rName, athena.WorkGroupStateDisabled),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAthenaWorkGroupExists(resourceName, &workgroup3),
					resource.TestCheckResourceAttr(resourceName, "state", athena.WorkGroupStateDisabled),
				),
			},
		},
	})
}

func TestAccAWSAthenaWorkGroup_Tags(t *testing.T) {
	var workgroup1, workgroup2, workgroup3 athena.WorkGroup
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_athena_workgroup.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAthenaWorkGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAthenaWorkGroupConfigTags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAthenaWorkGroupExists(resourceName, &workgroup1),
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
				Config: testAccAthenaWorkGroupConfigTags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAthenaWorkGroupExists(resourceName, &workgroup2),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccAthenaWorkGroupConfigTags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAthenaWorkGroupExists(resourceName, &workgroup3),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
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

		if isAWSErr(err, athena.ErrCodeInvalidRequestException, "is not found") {
			continue
		}

		if err != nil {
			return err
		}

		if resp.WorkGroup != nil {
			return fmt.Errorf("Athena WorkGroup (%s) found", rs.Primary.ID)
		}
	}
	return nil
}

func testAccCheckAWSAthenaWorkGroupExists(name string, workgroup *athena.WorkGroup) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		conn := testAccProvider.Meta().(*AWSClient).athenaconn

		input := &athena.GetWorkGroupInput{
			WorkGroup: aws.String(rs.Primary.ID),
		}

		output, err := conn.GetWorkGroup(input)

		if err != nil {
			return err
		}

		*workgroup = *output.WorkGroup

		return nil
	}
}

func testAccCheckAWSAthenaWorkGroupDisappears(workgroup *athena.WorkGroup) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).athenaconn

		input := &athena.DeleteWorkGroupInput{
			WorkGroup: workgroup.Name,
		}

		_, err := conn.DeleteWorkGroup(input)

		return err
	}
}

func testAccAthenaWorkGroupConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_athena_workgroup" "test" {
  name = %[1]q
}
`, rName)
}

func testAccAthenaWorkGroupConfigDescription(rName string, rDescription string) string {
	return fmt.Sprintf(`
resource "aws_athena_workgroup" "test" {
  description = %[2]q
  name        = %[1]q
}
`, rName, rDescription)
}

func testAccAthenaWorkGroupConfigBytesScannedCutoffPerQuery(rName string, rBytesScannedCutoffPerQuery string) string {
	return fmt.Sprintf(`
resource "aws_athena_workgroup" "test" {
  bytes_scanned_cutoff_per_query = %[2]s
  name                           = %[1]q
}
`, rName, rBytesScannedCutoffPerQuery)
}

func testAccAthenaWorkGroupConfigEnforceWorkgroupConfiguration(rName string, rEnforce string) string {
	return fmt.Sprintf(`
resource "aws_athena_workgroup" "test" {
  enforce_workgroup_configuration = %[2]s
  name                            = %[1]q
}
`, rName, rEnforce)
}

func testAccAthenaWorkGroupConfigPublishCloudWatchMetricsEnabled(rName string, rEnable string) string {
	return fmt.Sprintf(`
resource "aws_athena_workgroup" "test" {
  name                               = %[1]q
  publish_cloudwatch_metrics_enabled = %[2]s
}
`, rName, rEnable)
}

func testAccAthenaWorkGroupConfigOutputLocation(rName string, rOutputLocation string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test"{
  bucket        = %[2]q
  force_destroy = true
}

resource "aws_athena_workgroup" "test" {
  name            = %[1]q
  output_location = "s3://${aws_s3_bucket.test.bucket}/test/output"
}
`, rName, rOutputLocation)
}

func testAccAthenaWorkGroupConfigEncryptionS3(rName string, rEncryption string) string {
	return fmt.Sprintf(`
resource "aws_athena_workgroup" "test" {
  encryption_option = %[2]q
  name              = %[1]q
}
`, rName, rEncryption)
}

func testAccAthenaWorkGroupConfigEncryptionKms(rName string, rEncryption string) string {
	return fmt.Sprintf(`
resource "aws_kms_key" "test" {
  deletion_window_in_days = 7
  description             = "Terraform Acceptance Testing"
}

resource "aws_athena_workgroup" "test" {
  encryption_option = %[2]q
  kms_key           = "${aws_kms_key.test.arn}"
  name              = %[1]q
}
`, rName, rEncryption)
}

func testAccAthenaWorkGroupConfigState(rName, state string) string {
	return fmt.Sprintf(`
resource "aws_athena_workgroup" "test" {
  name  = %[1]q
  state = %[2]q
}
`, rName, state)
}

func testAccAthenaWorkGroupConfigTags1(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_athena_workgroup" "test" {
  name = %[1]q

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccAthenaWorkGroupConfigTags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_athena_workgroup" "test" {
  name = %[1]q

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}
