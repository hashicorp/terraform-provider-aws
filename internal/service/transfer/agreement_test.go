// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package transfer_test

import (
	"context"
	"fmt"
	"testing"

	awstypes "github.com/aws/aws-sdk-go-v2/service/transfer/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftransfer "github.com/hashicorp/terraform-provider-aws/internal/service/transfer"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccAgreement_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var conf awstypes.DescribedAgreement
	baseDirectory1 := "/DOC-EXAMPLE-BUCKET/home/mydirectory1"
	baseDirectory2 := "/DOC-EXAMPLE-BUCKET/home/mydirectory2"
	resourceName := "aws_transfer_agreement.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.TransferEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.TransferServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAgreementDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAgreementConfig_basic(rName, baseDirectory1),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAgreementExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "base_directory", baseDirectory1),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ""),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrStatus},
			},
			{
				Config: testAccAgreementConfig_basic(rName, baseDirectory2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAgreementExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "base_directory", baseDirectory2),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ""),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
				),
			},
		},
	})
}

func testAccAgreement_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var conf awstypes.DescribedAgreement
	resourceName := "aws_transfer_agreement.test"
	baseDirectory := "/DOC-EXAMPLE-BUCKET/home/mydirectory"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.TransferEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.TransferServiceID),
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
	var conf awstypes.DescribedAgreement
	baseDirectory := "/DOC-EXAMPLE-BUCKET/home/mydirectory"
	resourceName := "aws_transfer_agreement.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.TransferEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.TransferServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAgreementDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAgreementConfig_tags1(rName, baseDirectory, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAgreementExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrStatus},
			},
			{
				Config: testAccAgreementConfig_tags2(rName, baseDirectory, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAgreementExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccAgreementConfig_tags1(rName, baseDirectory, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAgreementExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func testAccCheckAgreementExists(ctx context.Context, n string, v *awstypes.DescribedAgreement) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).TransferClient(ctx)

		output, err := tftransfer.FindAgreementByTwoPartKey(ctx, conn, rs.Primary.Attributes["server_id"], rs.Primary.Attributes["agreement_id"])

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckAgreementDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).TransferClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_transfer_agreement" {
				continue
			}

			_, err := tftransfer.FindAgreementByTwoPartKey(ctx, conn, rs.Primary.Attributes["server_id"], rs.Primary.Attributes["agreement_id"])

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
  "Statement":[{
    "Sid":"AllowFullAccesstoS3",
    "Effect":"Allow",
    "Action":[
      "s3:*"
    ],
    "Resource":"*"
  }]
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
