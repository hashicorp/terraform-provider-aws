package datasync_test

import (
	"errors"
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/datasync"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfdatasync "github.com/hashicorp/terraform-provider-aws/internal/service/datasync"
)

func TestAccDataSyncLocationS3_basic(t *testing.T) {
	var locationS31 datasync.DescribeLocationS3Output
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	iamRoleResourceName := "aws_iam_role.test"
	resourceName := "aws_datasync_location_s3.test"
	s3BucketResourceName := "aws_s3_bucket.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, datasync.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckLocationS3Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccLocationS3Config(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLocationS3Exists(resourceName, &locationS31),
					resource.TestCheckResourceAttr(resourceName, "agent_arns.#", "0"),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "datasync", regexp.MustCompile(`location/loc-.+`)),
					resource.TestCheckResourceAttrPair(resourceName, "s3_bucket_arn", s3BucketResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "s3_config.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "s3_config.0.bucket_access_role_arn", iamRoleResourceName, "arn"),
					resource.TestCheckResourceAttrSet(resourceName, "s3_storage_class"),
					resource.TestCheckResourceAttr(resourceName, "subdirectory", "/test/"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestMatchResourceAttr(resourceName, "uri", regexp.MustCompile(`^s3://.+/`)),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"s3_bucket_arn"},
			},
		},
	})
}

func TestAccDataSyncLocationS3_storageClass(t *testing.T) {
	var locationS31 datasync.DescribeLocationS3Output
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	iamRoleResourceName := "aws_iam_role.test"
	resourceName := "aws_datasync_location_s3.test"
	s3BucketResourceName := "aws_s3_bucket.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, datasync.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckLocationS3Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccLocationS3StorageClassConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLocationS3Exists(resourceName, &locationS31),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "datasync", regexp.MustCompile(`location/loc-.+`)),
					resource.TestCheckResourceAttrPair(resourceName, "s3_bucket_arn", s3BucketResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "s3_config.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "s3_config.0.bucket_access_role_arn", iamRoleResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "subdirectory", "/test/"),
					resource.TestCheckResourceAttr(resourceName, "s3_storage_class", "STANDARD_IA"),
					resource.TestMatchResourceAttr(resourceName, "uri", regexp.MustCompile(`^s3://.+/`)),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"s3_bucket_arn"},
			},
		},
	})
}

func TestAccDataSyncLocationS3_disappears(t *testing.T) {
	var locationS31 datasync.DescribeLocationS3Output
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_datasync_location_s3.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, datasync.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckLocationS3Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccLocationS3Config(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLocationS3Exists(resourceName, &locationS31),
					acctest.CheckResourceDisappears(acctest.Provider, tfdatasync.ResourceLocationS3(), resourceName),
					acctest.CheckResourceDisappears(acctest.Provider, tfdatasync.ResourceLocationS3(), resourceName)),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccDataSyncLocationS3_tags(t *testing.T) {
	var locationS31, locationS32, locationS33 datasync.DescribeLocationS3Output
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_datasync_location_s3.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, datasync.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckLocationS3Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccLocationS3Tags1Config(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLocationS3Exists(resourceName, &locationS31),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"s3_bucket_arn"},
			},
			{
				Config: testAccLocationS3Tags2Config(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLocationS3Exists(resourceName, &locationS32),
					testAccCheckLocationS3NotRecreated(&locationS31, &locationS32),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccLocationS3Tags1Config(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLocationS3Exists(resourceName, &locationS33),
					testAccCheckLocationS3NotRecreated(&locationS32, &locationS33),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
		},
	})
}

func testAccCheckLocationS3Destroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).DataSyncConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_datasync_location_s3" {
			continue
		}

		input := &datasync.DescribeLocationS3Input{
			LocationArn: aws.String(rs.Primary.ID),
		}

		_, err := conn.DescribeLocationS3(input)

		if tfawserr.ErrMessageContains(err, "InvalidRequestException", "not found") {
			return nil
		}

		if err != nil {
			return err
		}
	}

	return nil
}

func testAccCheckLocationS3Exists(resourceName string, locationS3 *datasync.DescribeLocationS3Output) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).DataSyncConn
		input := &datasync.DescribeLocationS3Input{
			LocationArn: aws.String(rs.Primary.ID),
		}

		output, err := conn.DescribeLocationS3(input)

		if err != nil {
			return err
		}

		if output == nil {
			return fmt.Errorf("Location %q does not exist", rs.Primary.ID)
		}

		*locationS3 = *output

		return nil
	}
}

func testAccCheckLocationS3NotRecreated(i, j *datasync.DescribeLocationS3Output) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if !aws.TimeValue(i.CreationTime).Equal(aws.TimeValue(j.CreationTime)) {
			return errors.New("DataSync Location S3 was recreated")
		}

		return nil
	}
}

func testAccLocationS3BaseConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = <<POLICY
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Service": "datasync.amazonaws.com"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
POLICY
}

resource "aws_iam_role_policy" "test" {
  role   = aws_iam_role.test.id
  policy = <<POLICY
{
  "Version": "2012-10-17",
  "Statement": [{
    "Action": [
      "s3:*"
    ],
    "Effect": "Allow",
    "Resource": [
      "${aws_s3_bucket.test.arn}",
      "${aws_s3_bucket.test.arn}/*"
    ]
  }]
}
POLICY
}

resource "aws_s3_bucket" "test" {
  bucket        = %[1]q
  force_destroy = true
}
`, rName)
}

func testAccLocationS3Config(rName string) string {
	return testAccLocationS3BaseConfig(rName) + `
resource "aws_datasync_location_s3" "test" {
  s3_bucket_arn = aws_s3_bucket.test.arn
  subdirectory  = "/test"

  s3_config {
    bucket_access_role_arn = aws_iam_role.test.arn
  }
}
`
}

func testAccLocationS3StorageClassConfig(rName string) string {
	return testAccLocationS3BaseConfig(rName) + `
resource "aws_datasync_location_s3" "test" {
  s3_bucket_arn    = aws_s3_bucket.test.arn
  subdirectory     = "/test"
  s3_storage_class = "STANDARD_IA"

  s3_config {
    bucket_access_role_arn = aws_iam_role.test.arn
  }
}
`
}

func testAccLocationS3Tags1Config(rName, key1, value1 string) string {
	return testAccLocationS3BaseConfig(rName) + fmt.Sprintf(`
resource "aws_datasync_location_s3" "test" {
  s3_bucket_arn = aws_s3_bucket.test.arn
  subdirectory  = "/test"

  s3_config {
    bucket_access_role_arn = aws_iam_role.test.arn
  }

  tags = {
    %q = %q
  }
}
`, key1, value1)
}

func testAccLocationS3Tags2Config(rName, key1, value1, key2, value2 string) string {
	return testAccLocationS3BaseConfig(rName) + fmt.Sprintf(`
resource "aws_datasync_location_s3" "test" {
  s3_bucket_arn = aws_s3_bucket.test.arn
  subdirectory  = "/test"

  s3_config {
    bucket_access_role_arn = aws_iam_role.test.arn
  }

  tags = {
    %q = %q
    %q = %q
  }
}
`, key1, value1, key2, value2)
}
