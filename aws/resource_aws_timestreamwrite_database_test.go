package aws

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/timestreamwrite"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccAWSTimestreamWriteDatabase_basic(t *testing.T) {
	resourceName := "aws_timestreamwrite_database.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSTimestreamWriteDatabaseDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSTimestreamWriteDatabaseConfigNoTags(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSTimestreamWriteDatabaseExists(resourceName),
					testAccCheckResourceAttrRegionalARN(resourceName, "arn", "timestream", fmt.Sprintf("database/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "database_name", rName),
					testAccMatchResourceAttrRegionalARN(resourceName, "kms_key_id", "kms", regexp.MustCompile(`key/.+`)),
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

func TestAccAWSTimestreamWriteDatabase_kmsKey(t *testing.T) {
	resourceName := "aws_timestreamwrite_database.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")
	kmsResourceName := "aws_kms_key.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSTimestreamWriteDatabaseDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSTimestreamWriteDatabaseConfigKmsKey(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSTimestreamWriteDatabaseExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "database_name", rName),
					resource.TestCheckResourceAttrPair(resourceName, "kms_key_id", kmsResourceName, "arn"),
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

func TestAccAWSTimestreamWriteDatabase_Tags(t *testing.T) {
	resourceName := "aws_timestreamwrite_database.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSTimestreamWriteDatabaseDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSTimestreamWriteDatabaseConfigTags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSTimestreamWriteDatabaseExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				Config: testAccAWSTimestreamWriteDatabaseConfigTags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSTimestreamWriteDatabaseExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccAWSTimestreamWriteDatabaseConfigTags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSTimestreamWriteDatabaseExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
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

func testAccCheckAWSTimestreamWriteDatabaseDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).timestreamwriteconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_timestreamwrite_database" {
			continue
		}

		_, err := conn.DescribeDatabase(&timestreamwrite.DescribeDatabaseInput{
			DatabaseName: aws.String(rs.Primary.ID),
		})

		if isAWSErr(err, timestreamwrite.ErrCodeResourceNotFoundException, "") {
			continue
		}

		if err == nil {
			return fmt.Errorf("Timestream Database (%s) still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccCheckAWSTimestreamWriteDatabaseExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Timestream Database ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).timestreamwriteconn

		_, err := conn.DescribeDatabase(&timestreamwrite.DescribeDatabaseInput{
			DatabaseName: aws.String(rs.Primary.ID),
		})

		if err != nil {
			return err
		}

		return nil
	}
}

func testAccAWSTimestreamWriteDatabaseConfigNoTags(rName string) string {
	return fmt.Sprintf(`
resource "aws_timestreamwrite_database" "test" {
  database_name = %[1]q
}
`, rName)
}

func testAccAWSTimestreamWriteDatabaseConfigTags1(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_timestreamwrite_database" "test" {
  database_name = %[1]q

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccAWSTimestreamWriteDatabaseConfigTags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_timestreamwrite_database" "test" {
  database_name = %[1]q

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}

func testAccAWSTimestreamWriteDatabaseConfigKmsKey(rName string) string {
	return fmt.Sprintf(`
resource "aws_kms_key" "test" {
  description = "Terraform acc test"

  policy = <<POLICY
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "AWS": "*"
      },
      "Action": "kms:*",
      "Resource": "*"
    }
  ]
}
POLICY
}

resource "aws_timestreamwrite_database" "test" {
  database_name = %[1]q
  kms_key_id    = aws_kms_key.test.arn
}
`, rName)
}
