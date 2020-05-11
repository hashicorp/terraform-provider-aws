package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/lakeformation"

	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestAccAWSLakeFormationResource_basic(t *testing.T) {
	bName := acctest.RandomWithPrefix("lakeformation-test-bucket")
	resourceName := "aws_lakeformation_resource.test"
	bucketName := "aws_s3_bucket.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSLakeFormationResourceDeregister,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLakeFormationResourceConfig_basic(bName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(bucketName, "arn", resourceName, "resource_arn"),
					resource.TestCheckResourceAttr(resourceName, "use_service_linked_role", "true"),
					resource.TestCheckResourceAttrSet(resourceName, "role_arn"),
				),
			},
		},
	})
}

func TestAccAWSLakeFormationResource_withRole(t *testing.T) {
	bName := acctest.RandomWithPrefix("lakeformation-test-bucket")
	resourceName := "aws_lakeformation_resource.test"
	bucketName := "aws_s3_bucket.test"
	roleName := "data.aws_iam_role.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSLakeFormationResourceDeregister,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLakeFormationResourceConfig_withRole(bName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(bucketName, "arn", resourceName, "resource_arn"),
					resource.TestCheckResourceAttrPair(roleName, "arn", resourceName, "role_arn"),
					resource.TestCheckResourceAttr(resourceName, "use_service_linked_role", "false"),
				),
			},
		},
	})
}

func TestAccAWSLakeFormationResource_update(t *testing.T) {
	bName := acctest.RandomWithPrefix("lakeformation-test-bucket")
	resourceName := "aws_lakeformation_resource.test"
	bucketName := "aws_s3_bucket.test"
	roleName := "data.aws_iam_role.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSLakeFormationResourceDeregister,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLakeFormationResourceConfig_basic(bName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(bucketName, "arn", resourceName, "resource_arn"),
					resource.TestCheckResourceAttr(resourceName, "use_service_linked_role", "true"),
					resource.TestCheckResourceAttrSet(resourceName, "role_arn"),
				),
			},
			{
				Config: testAccAWSLakeFormationResourceConfig_withRole(bName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(bucketName, "arn", resourceName, "resource_arn"),
					resource.TestCheckResourceAttrPair(roleName, "arn", resourceName, "role_arn"),
					resource.TestCheckResourceAttr(resourceName, "use_service_linked_role", "false"),
				),
			},
		},
	})
}

func testAccAWSLakeFormationResourceConfig_basic(bName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_lakeformation_resource" "test" {
  resource_arn            = "${aws_s3_bucket.test.arn}"
  use_service_linked_role = true
}
`, bName)
}

func testAccAWSLakeFormationResourceConfig_withRole(bName string) string {
	return fmt.Sprintf(`
data "aws_iam_role" "test" {
  name = "AWSServiceRoleForLakeFormationDataAccess"
}

resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_lakeformation_resource" "test" {
  resource_arn            = "${aws_s3_bucket.test.arn}"
  role_arn                = "${data.aws_iam_role.test.arn}"
  use_service_linked_role = false
}
`, bName)
}

func testAccCheckAWSLakeFormationResourceDeregister(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).lakeformationconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_lakeformation_resource" {
			continue
		}

		resourceArn := rs.Primary.Attributes["resource_arn"]

		input := &lakeformation.DescribeResourceInput{
			ResourceArn: aws.String(resourceArn),
		}

		_, err := conn.DescribeResource(input)
		if err == nil {
			return fmt.Errorf("Resource still registered: %s", resourceArn)
		}
		if !isLakeFormationResourceNotFoundErr(err) {
			return err
		}
	}

	return nil
}

func isLakeFormationResourceNotFoundErr(err error) bool {
	return isAWSErr(
		err,
		"EntityNotFoundException",
		"Entity not found")
}
