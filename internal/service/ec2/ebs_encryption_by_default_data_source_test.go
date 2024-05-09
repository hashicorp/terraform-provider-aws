// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2_test

import (
	"context"
	"fmt"
	"strconv"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccEC2EBSEncryptionByDefaultDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccEBSEncryptionByDefaultDataSourceConfig_basic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEBSEncryptionByDefaultDataSource(ctx, "data.aws_ebs_encryption_by_default.current"),
				),
			},
		},
	})
}

func testAccCheckEBSEncryptionByDefaultDataSource(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Client(ctx)

		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		actual, err := conn.GetEbsEncryptionByDefault(ctx, &ec2.GetEbsEncryptionByDefaultInput{})
		if err != nil {
			return fmt.Errorf("Error reading default EBS encryption toggle: %q", err)
		}

		attr, _ := strconv.ParseBool(rs.Primary.Attributes[names.AttrEnabled])

		if attr != aws.ToBool(actual.EbsEncryptionByDefault) {
			return fmt.Errorf("EBS encryption by default is not in expected state (%t)", aws.ToBool(actual.EbsEncryptionByDefault))
		}

		return nil
	}
}

const testAccEBSEncryptionByDefaultDataSourceConfig_basic = `
data "aws_ebs_encryption_by_default" "current" {}
`
