package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	kmsfinder "github.com/hashicorp/terraform-provider-aws/aws/internal/service/kms/finder"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/provider"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
)

func TestAccAWSEBSDefaultKmsKey_basic(t *testing.T) {
	resourceName := "aws_ebs_default_kms_key.test"
	resourceNameKey := "aws_kms_key.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ec2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckAwsEbsDefaultKmsKeyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsEbsDefaultKmsKeyConfig_basic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEbsDefaultKmsKey(resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "key_arn", resourceNameKey, "arn"),
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

func testAccCheckAwsEbsDefaultKmsKeyDestroy(s *terraform.State) error {
	arn, err := testAccAwsEbsDefaultKmsKeyAwsManagedDefaultKey()
	if err != nil {
		return err
	}

	conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn

	resp, err := conn.GetEbsDefaultKmsKeyId(&ec2.GetEbsDefaultKmsKeyIdInput{})
	if err != nil {
		return err
	}

	// Verify that the default key is now the account's AWS-managed default CMK.
	if aws.StringValue(resp.KmsKeyId) != arn.String() {
		return fmt.Errorf("Default CMK (%s) is not the account's AWS-managed default CMK (%s)", aws.StringValue(resp.KmsKeyId), arn.String())
	}

	return nil
}

func testAccCheckEbsDefaultKmsKey(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		arn, err := testAccAwsEbsDefaultKmsKeyAwsManagedDefaultKey()
		if err != nil {
			return err
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn

		resp, err := conn.GetEbsDefaultKmsKeyId(&ec2.GetEbsDefaultKmsKeyIdInput{})
		if err != nil {
			return err
		}

		// Verify that the default key is not the account's AWS-managed default CMK.
		if aws.StringValue(resp.KmsKeyId) == arn.String() {
			return fmt.Errorf("Default CMK (%s) is the account's AWS-managed default CMK (%s)", aws.StringValue(resp.KmsKeyId), arn.String())
		}

		return nil
	}
}

// testAccAwsEbsDefaultKmsKeyAwsManagedDefaultKey returns' the account's AWS-managed default CMK.
func testAccAwsEbsDefaultKmsKeyAwsManagedDefaultKey() (*arn.ARN, error) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).KMSConn

	alias, err := kmsfinder.AliasByName(conn, "alias/aws/ebs")
	if err != nil {
		return nil, err
	}

	aliasARN, err := arn.Parse(aws.StringValue(alias.AliasArn))
	if err != nil {
		return nil, err
	}

	arn := arn.ARN{
		Partition: aliasARN.Partition,
		Service:   aliasARN.Service,
		Region:    aliasARN.Region,
		AccountID: aliasARN.AccountID,
		Resource:  fmt.Sprintf("key/%s", aws.StringValue(alias.TargetKeyId)),
	}

	return &arn, nil
}

const testAccAwsEbsDefaultKmsKeyConfig_basic = `
resource "aws_kms_key" "test" {}

resource "aws_ebs_default_kms_key" "test" {
  key_arn = aws_kms_key.test.arn
}
`
