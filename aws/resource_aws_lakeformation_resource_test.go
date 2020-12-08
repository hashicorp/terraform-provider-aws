package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/lakeformation"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccAWSLakeFormationResource_basic(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_lakeformation_resource.test"
	bucketName := "aws_s3_bucket.test"
	roleName := "aws_iam_role.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSLakeFormationResourceDeregister,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLakeFormationResourceConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, "resource_arn", bucketName, "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "role_arn", roleName, "arn"),
					resource.TestCheckResourceAttrSet(resourceName, "last_modified"),
				),
			},
		},
	})
}

func TestAccAWSLakeFormationResource_withRole(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_lakeformation_resource.test"
	bucketName := "aws_s3_bucket.test"
	roleName := "data.aws_iam_role.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSLakeFormationResourceDeregister,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLakeFormationResourceConfig_withRole(rName),
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
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_lakeformation_resource.test"
	bucketName := "aws_s3_bucket.test"
	roleName := "data.aws_iam_role.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSLakeFormationResourceDeregister,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLakeFormationResourceConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(bucketName, "arn", resourceName, "resource_arn"),
					resource.TestCheckResourceAttr(resourceName, "use_service_linked_role", "true"),
					resource.TestCheckResourceAttrSet(resourceName, "role_arn"),
				),
			},
			{
				Config: testAccAWSLakeFormationResourceConfig_withRole(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(bucketName, "arn", resourceName, "resource_arn"),
					resource.TestCheckResourceAttrPair(roleName, "arn", resourceName, "role_arn"),
					resource.TestCheckResourceAttr(resourceName, "use_service_linked_role", "false"),
				),
			},
		},
	})
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

func testAccAWSLakeFormationResourceConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_iam_role" "test" {
  name = %[1]q
  path = "/test/"

  assume_role_policy = <<EOF
{
	"Version": "2012-10-17",
	"Statement": [
		{
			"Action": "sts:AssumeRole",
			"Principal": {
				"Service": "s3.amazonaws.com"
			},
			"Effect": "Allow",
			"Sid": ""
		}
	]
}
EOF
}

data "aws_partition" "current" {}

resource "aws_iam_role_policy" "test" {
  name = %[1]q
  role = aws_iam_role.test.id

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "s3:GetBucketLocation",
        "s3:ListAllMyBuckets",
        "s3:GetObjectVersion",
        "s3:GetBucketAcl",
        "s3:GetObject",
        "s3:GetObjectACL",
        "s3:PutObject",
        "s3:PutObjectAcl",
      ],
      "Resource": [
		"arn:${data.aws_partition.current.partition}:s3:::${aws_s3_bucket.test.id}/*",
		"arn:${data.aws_partition.current.partition}:s3:::${aws_s3_bucket.test.id}",
      ],
    }
  ]
}
EOF
}

resource "aws_lakeformation_resource" "test" {
  resource_arn = aws_s3_bucket.test.arn
  role_arn     = aws_iam_role.test.arn
}
`, rName)
}

func testAccAWSLakeFormationResourceConfig_serviceLinkedRole(rName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_lakeformation_resource" "test" {
  resource_arn            = "${aws_s3_bucket.test.arn}"
  use_service_linked_role = true
}
`, rName)
}

func testAccAWSLakeFormationResourceConfig_withRole(rName string) string {
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
`, rName)
}
