// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfkms "github.com/hashicorp/terraform-provider-aws/internal/service/kms"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccEC2EBSDefaultKMSKey_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_ebs_default_kms_key.test"
	resourceNameKey := "aws_kms_key.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEBSDefaultKMSKeyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEBSDefaultKMSKeyConfig_basic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEBSDefaultKMSKey(ctx, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "key_arn", resourceNameKey, names.AttrARN),
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

func testAccCheckEBSDefaultKMSKeyDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		arn, err := testAccEBSManagedDefaultKey(ctx)
		if err != nil {
			return err
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Client(ctx)

		resp, err := conn.GetEbsDefaultKmsKeyId(ctx, &ec2.GetEbsDefaultKmsKeyIdInput{})
		if err != nil {
			return err
		}

		// Verify that the default key is now the account's AWS-managed default CMK.
		if aws.ToString(resp.KmsKeyId) != arn.String() {
			return fmt.Errorf("Default CMK (%s) is not the account's AWS-managed default CMK (%s)", aws.ToString(resp.KmsKeyId), arn.String())
		}

		return nil
	}
}

func testAccCheckEBSDefaultKMSKey(ctx context.Context, name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		arn, err := testAccEBSManagedDefaultKey(ctx)
		if err != nil {
			return err
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Client(ctx)

		resp, err := conn.GetEbsDefaultKmsKeyId(ctx, &ec2.GetEbsDefaultKmsKeyIdInput{})
		if err != nil {
			return err
		}

		// Verify that the default key is not the account's AWS-managed default CMK.
		if aws.ToString(resp.KmsKeyId) == arn.String() {
			return fmt.Errorf("Default CMK (%s) is the account's AWS-managed default CMK (%s)", aws.ToString(resp.KmsKeyId), arn.String())
		}

		return nil
	}
}

// testAccEBSManagedDefaultKey returns' the account's AWS-managed default CMK.
func testAccEBSManagedDefaultKey(ctx context.Context) (*arn.ARN, error) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).KMSClient(ctx)

	alias, err := tfkms.FindAliasByName(ctx, conn, "alias/aws/ebs")
	if err != nil {
		return nil, err
	}

	aliasARN, err := arn.Parse(aws.ToString(alias.AliasArn))
	if err != nil {
		return nil, err
	}

	arn := arn.ARN{
		Partition: aliasARN.Partition,
		Service:   aliasARN.Service,
		Region:    aliasARN.Region,
		AccountID: aliasARN.AccountID,
		Resource:  fmt.Sprintf("key/%s", aws.ToString(alias.TargetKeyId)),
	}

	return &arn, nil
}

const testAccEBSDefaultKMSKeyConfig_basic = `
resource "aws_kms_key" "test" {}

resource "aws_ebs_default_kms_key" "test" {
  key_arn = aws_kms_key.test.arn
}
`
