package lakeformation_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/lakeformation"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tflakeformation "github.com/hashicorp/terraform-provider-aws/internal/service/lakeformation"
)

func TestAccLakeFormationResource_basic(t *testing.T) {
	bucketName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	roleName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceAddr := "aws_lakeformation_resource.test"
	bucketAddr := "aws_s3_bucket.test"
	roleAddr := "aws_iam_role.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(lakeformation.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, lakeformation.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceConfig_basic(bucketName, roleName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResourceExists(resourceAddr),
					resource.TestCheckResourceAttrPair(resourceAddr, "role_arn", roleAddr, "arn"),
					resource.TestCheckResourceAttrPair(resourceAddr, "arn", bucketAddr, "arn"),
				),
			},
		},
	})
}

func TestAccLakeFormationResource_disappears(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lakeformation_resource.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(lakeformation.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, lakeformation.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceConfig_basic(rName, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResourceExists(resourceName),
					acctest.CheckResourceDisappears(acctest.Provider, tflakeformation.ResourceResource(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccLakeFormationResource_serviceLinkedRole(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceAddr := "aws_lakeformation_resource.test"
	bucketAddr := "aws_s3_bucket.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(lakeformation.EndpointsID, t)
			acctest.PreCheckIAMServiceLinkedRole(t, "/aws-service-role/lakeformation.amazonaws.com")
		},
		ErrorCheck:        acctest.ErrorCheck(t, lakeformation.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceConfig_serviceLinkedRole(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResourceExists(resourceAddr),
					resource.TestCheckResourceAttrPair(resourceAddr, "arn", bucketAddr, "arn"),
					acctest.CheckResourceAttrGlobalARN(resourceAddr, "role_arn", "iam", "role/aws-service-role/lakeformation.amazonaws.com/AWSServiceRoleForLakeFormationDataAccess"),
				),
			},
		},
	})
}

func TestAccLakeFormationResource_updateRoleToRole(t *testing.T) {
	bucketName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	roleName1 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	roleName2 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceAddr := "aws_lakeformation_resource.test"
	bucketAddr := "aws_s3_bucket.test"
	roleAddr := "aws_iam_role.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(lakeformation.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, lakeformation.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceConfig_basic(bucketName, roleName1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResourceExists(resourceAddr),
					resource.TestCheckResourceAttrPair(resourceAddr, "role_arn", roleAddr, "arn"),
					resource.TestCheckResourceAttrPair(resourceAddr, "arn", bucketAddr, "arn"),
				),
			},
			{
				Config: testAccResourceConfig_basic(bucketName, roleName2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResourceExists(resourceAddr),
					resource.TestCheckResourceAttrPair(resourceAddr, "role_arn", roleAddr, "arn"),
					resource.TestCheckResourceAttrPair(resourceAddr, "arn", bucketAddr, "arn"),
				),
			},
		},
	})
}

func TestAccLakeFormationResource_updateSLRToRole(t *testing.T) {
	bucketName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	roleName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceAddr := "aws_lakeformation_resource.test"
	bucketAddr := "aws_s3_bucket.test"
	roleAddr := "aws_iam_role.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(lakeformation.EndpointsID, t)
			acctest.PreCheckIAMServiceLinkedRole(t, "/aws-service-role/lakeformation.amazonaws.com")
		},
		ErrorCheck:        acctest.ErrorCheck(t, lakeformation.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceConfig_serviceLinkedRole(bucketName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResourceExists(resourceAddr),
					resource.TestCheckResourceAttrPair(resourceAddr, "arn", bucketAddr, "arn"),
					acctest.CheckResourceAttrGlobalARN(resourceAddr, "role_arn", "iam", "role/aws-service-role/lakeformation.amazonaws.com/AWSServiceRoleForLakeFormationDataAccess"),
				),
			},
			{
				Config: testAccResourceConfig_basic(bucketName, roleName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResourceExists(resourceAddr),
					resource.TestCheckResourceAttrPair(resourceAddr, "role_arn", roleAddr, "arn"),
					resource.TestCheckResourceAttrPair(resourceAddr, "arn", bucketAddr, "arn"),
				),
			},
		},
	})
}

// AWS does not support changing from an IAM role to an SLR. No error is thrown
// but the registration is not changed (the IAM role continues in the registration).
//
// func TestAccLakeFormationResource_updateRoleToSLR(t *testing.T) {

func testAccCheckResourceDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).LakeFormationConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_lakeformation_resource" {
			continue
		}

		resourceArn := rs.Primary.Attributes["arn"]

		input := &lakeformation.DescribeResourceInput{
			ResourceArn: aws.String(resourceArn),
		}

		_, err := conn.DescribeResource(input)
		if err == nil {
			return fmt.Errorf("resource still registered: %s", resourceArn)
		}
		if !isResourceNotFoundErr(err) {
			return err
		}
	}

	return nil
}

func testAccCheckResourceExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource not found: %s", resourceName)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).LakeFormationConn

		input := &lakeformation.DescribeResourceInput{
			ResourceArn: aws.String(rs.Primary.ID),
		}

		_, err := conn.DescribeResource(input)

		if err != nil {
			return fmt.Errorf("error getting Lake Formation resource (%s): %w", rs.Primary.ID, err)
		}

		return nil
	}
}

func isResourceNotFoundErr(err error) bool {
	return tfawserr.ErrMessageContains(
		err,
		"EntityNotFoundException",
		"Entity not found")

}

func testAccResourceConfig_basic(bucket, role string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_iam_role" "test" {
  name = %[2]q
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
  name = %[2]q
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
        "s3:PutObjectAcl"
      ],
      "Resource": [
        "arn:${data.aws_partition.current.partition}:s3:::${aws_s3_bucket.test.id}/*",
        "arn:${data.aws_partition.current.partition}:s3:::${aws_s3_bucket.test.id}"
      ]
    }
  ]
}
EOF
}

resource "aws_lakeformation_resource" "test" {
  arn      = aws_s3_bucket.test.arn
  role_arn = aws_iam_role.test.arn
}
`, bucket, role)
}

func testAccResourceConfig_serviceLinkedRole(rName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_lakeformation_resource" "test" {
  arn = aws_s3_bucket.test.arn
}
`, rName)
}
