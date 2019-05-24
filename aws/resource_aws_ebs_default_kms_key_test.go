package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAWSEBSDefaultKmsKey_basic(t *testing.T) {
	resourceName := "aws_ebs_default_kms_key.test"
	resourceNameKey1 := "aws_kms_key.test1"
	resourceNameKey2 := "aws_kms_key.test2"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsEbsDefaultKmsKeyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsEbsDefaultKmsKeyConfig_basic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEbsDefaultKmsKey(resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "key_id", resourceNameKey1, "arn"),
				),
			},
			{
				Config: testAccAwsEbsDefaultKmsKeyConfig_updated,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEbsDefaultKmsKey(resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "key_id", resourceNameKey2, "arn"),
				),
			},
		},
	})
}

func testAccCheckAwsEbsDefaultKmsKeyDestroy(s *terraform.State) error {
	ec2conn := testAccProvider.Meta().(*AWSClient).ec2conn
	kmsconn := testAccProvider.Meta().(*AWSClient).kmsconn

	alias, err := findKmsAliasByName(kmsconn, "alias/aws/ebs", nil)
	if err != nil {
		return err
	}

	aliasARN, err := arn.Parse(aws.StringValue(alias.AliasArn))
	if err != nil {
		return err
	}

	arn := arn.ARN{
		Partition: aliasARN.Partition,
		Service:   aliasARN.Service,
		Region:    aliasARN.Region,
		AccountID: aliasARN.AccountID,
		Resource:  fmt.Sprintf("key/%s", aws.StringValue(alias.TargetKeyId)),
	}

	resp, err := ec2conn.GetEbsDefaultKmsKeyId(&ec2.GetEbsDefaultKmsKeyIdInput{})
	if err != nil {
		return err
	}

	if aws.StringValue(resp.KmsKeyId) != arn.String() {
		return fmt.Errorf("Default CMK (%s) is not the account's AWS-managed default CMK (%s)", aws.StringValue(resp.KmsKeyId), arn.String())
	}

	return nil
}

func testAccCheckEbsDefaultKmsKey(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		return nil
	}
}

const testAccAwsEbsDefaultKmsKeyConfigBase = `
resource "aws_kms_key" "test1" {}

resource "aws_kms_key" "test2" {}
`

const testAccAwsEbsDefaultKmsKeyConfig_basic = testAccAwsEbsDefaultKmsKeyConfigBase + `
resource "aws_ebs_default_kms_key" "test" {
  key_id = "${aws_kms_key.test1.arn}"
}
`

const testAccAwsEbsDefaultKmsKeyConfig_updated = testAccAwsEbsDefaultKmsKeyConfigBase + `
resource "aws_ebs_default_kms_key" "test" {
  key_id = "${aws_kms_key.test2.arn}"
}
`
