// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package transfer_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/transfer"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftransfer "github.com/hashicorp/terraform-provider-aws/internal/service/transfer"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func testAccAgreement_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var conf transfer.DescribedAgreement
	baseDirectory1 := "/DOC-EXAMPLE-BUCKET/home/mydirectory1"
	baseDirectory2 := "/DOC-EXAMPLE-BUCKET/home/mydirectory2"
	resourceName := "aws_transfer_agreement.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, transfer.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAgreementDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAgreementConfig_basic(rName, baseDirectory1),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAgreementExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "base_directory", baseDirectory1),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"status"},
			},
			{
				Config: testAccAgreementConfig_basic(rName, baseDirectory2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAgreementExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "base_directory", baseDirectory2),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
		},
	})
}

func testAccAgreement_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var conf transfer.DescribedAgreement
	resourceName := "aws_transfer_agreement.test"
	baseDirectory := "/DOC-EXAMPLE-BUCKET/home/mydirectory"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, transfer.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAgreementDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAgreementConfig_basic(rName, baseDirectory),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAgreementExists(ctx, resourceName, &conf),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tftransfer.ResourceAgreement(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccAgreement_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var conf transfer.DescribedAgreement
	baseDirectory := "/DOC-EXAMPLE-BUCKET/home/mydirectory"
	resourceName := "aws_transfer_agreement.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, transfer.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAgreementDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAgreementConfig_tags1(rName, baseDirectory, "key1", "value1"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAgreementExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"status"},
			},
			{
				Config: testAccAgreementConfig_tags2(rName, baseDirectory, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAgreementExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccAgreementConfig_tags1(rName, baseDirectory, "key2", "value2"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAgreementExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func testAccCheckAgreementExists(ctx context.Context, n string, v *transfer.DescribedAgreement) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Transfer Agreement ID is set")
		}

		serverID, agreementID, err := tftransfer.AgreementParseResourceID(rs.Primary.ID)

		if err != nil {
			return err
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).TransferConn(ctx)

		output, err := tftransfer.FindAgreementByTwoPartKey(ctx, conn, serverID, agreementID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckAgreementDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).TransferConn(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_transfer_agreement" {
				continue
			}

			serverID, agreementID, err := tftransfer.AgreementParseResourceID(rs.Primary.ID)

			if err != nil {
				return err
			}

			_, err = tftransfer.FindAgreementByTwoPartKey(ctx, conn, serverID, agreementID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Transfer Agreement %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccAgreementConfig_base(rName string) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "test" {
  name               = %[1]q
  assume_role_policy = <<EOF
  {
	"Version": "2012-10-17",
	"Statement": [{
	  "Effect": "Allow",
	  "Principal": {
		"Service": "transfer.amazonaws.com"
	  },
	  "Action": "sts:AssumeRole"
	}]
  }
EOF
}

resource "aws_iam_role_policy" "test" {
  name = %[1]q
  role = aws_iam_role.test.id

  policy = <<POLICY
{
	 "Version":"2012-10-17",
	 "Statement":[
		{
		   "Sid":"AllowFullAccesstoS3",
		   "Effect":"Allow",
		   "Action":[
			  "s3:*"
		   ],
		   "Resource":"*"
		}
	 ]
}
POLICY
}

resource "aws_transfer_profile" "local" {
  as2_id       = %[1]q
  profile_type = "LOCAL"
}

resource "aws_transfer_profile" "partner" {
  as2_id       = %[1]q
  profile_type = "PARTNER"
}

resource "aws_transfer_server" "test" {}
`, rName)
}

func testAccAgreementConfig_basic(rName, baseDirectory string) string {
	return acctest.ConfigCompose(testAccAgreementConfig_base(rName), fmt.Sprintf(`
resource "aws_transfer_agreement" "test" {
  access_role        = aws_iam_role.test.arn
  base_directory     = %[1]q
  local_profile_id   = aws_transfer_profile.local.profile_id
  partner_profile_id = aws_transfer_profile.partner.profile_id
  server_id          = aws_transfer_server.test.id
}
`, baseDirectory))
}

func testAccAgreementConfig_tags1(rName, baseDirectory, tagKey1, tagValue1 string) string {
	return acctest.ConfigCompose(testAccAgreementConfig_base(rName), fmt.Sprintf(`
resource "aws_transfer_agreement" "test" {
  access_role        = aws_iam_role.test.arn
  base_directory     = %[1]q
  local_profile_id   = aws_transfer_profile.local.profile_id
  partner_profile_id = aws_transfer_profile.partner.profile_id
  server_id          = aws_transfer_server.test.id

  tags = {
    %[2]q = %[3]q
  }
}
`, baseDirectory, tagKey1, tagValue1))
}

func testAccAgreementConfig_tags2(rName, baseDirectory, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return acctest.ConfigCompose(testAccAgreementConfig_base(rName), fmt.Sprintf(`
resource "aws_transfer_agreement" "test" {
  access_role        = aws_iam_role.test.arn
  base_directory     = %[1]q
  local_profile_id   = aws_transfer_profile.local.profile_id
  partner_profile_id = aws_transfer_profile.partner.profile_id
  server_id          = aws_transfer_server.test.id

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, baseDirectory, tagKey1, tagValue1, tagKey2, tagValue2))
}
