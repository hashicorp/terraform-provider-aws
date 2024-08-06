// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ram_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ram/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfram "github.com/hashicorp/terraform-provider-aws/internal/service/ram"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func init() {
	acctest.RegisterServiceErrorCheckFunc(names.RAMServiceID, testAccErrorCheckSkip)
}

func testAccErrorCheckSkip(t *testing.T) resource.ErrorCheckFunc {
	return acctest.ErrorCheckSkipMessagesContaining(t,
		"this resource is not necessary",
	)
}

func TestAccRAMResourceShareAccepter_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_ram_resource_share_accepter.test"
	principalAssociationResourceName := "aws_ram_principal_association.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckAlternateAccount(t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.RAMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesAlternate(ctx, t),
		CheckDestroy:             testAccCheckResourceShareAccepterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceShareAccepterConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResourceShareAccepterExists(ctx, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "share_arn", principalAssociationResourceName, "resource_share_arn"),
					acctest.MatchResourceAttrRegionalARNAccountID(resourceName, "invitation_arn", "ram", `\d{12}`, regexache.MustCompile(fmt.Sprintf("resource-share-invitation/%s$", verify.UUIDRegexPattern))),
					resource.TestMatchResourceAttr(resourceName, "share_id", regexache.MustCompile(fmt.Sprintf(`^rs-%s$`, verify.UUIDRegexPattern))),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, string(awstypes.ResourceShareStatusActive)),
					acctest.CheckResourceAttrAccountID(resourceName, "receiver_account_id"),
					resource.TestMatchResourceAttr(resourceName, "sender_account_id", regexache.MustCompile(`\d{12}`)),
					resource.TestCheckResourceAttr(resourceName, "share_name", rName),
					resource.TestCheckResourceAttr(resourceName, "resources.%", acctest.Ct0),
				),
			},
			{
				Config:            testAccResourceShareAccepterConfig_basic(rName),
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccRAMResourceShareAccepter_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_ram_resource_share_accepter.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckAlternateAccount(t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.RAMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesAlternate(ctx, t),
		CheckDestroy:             testAccCheckResourceShareAccepterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceShareAccepterConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResourceShareAccepterExists(ctx, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfram.ResourceResourceShareAccepter(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccRAMResourceShareAccepter_resourceAssociation(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_ram_resource_share_accepter.test"
	principalAssociationResourceName := "aws_ram_principal_association.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckAlternateAccount(t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.RAMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesAlternate(ctx, t),
		CheckDestroy:             testAccCheckResourceShareAccepterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceShareAccepterConfig_association(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResourceShareAccepterExists(ctx, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "share_arn", principalAssociationResourceName, "resource_share_arn"),
					acctest.MatchResourceAttrRegionalARNAccountID(resourceName, "invitation_arn", "ram", `\d{12}`, regexache.MustCompile(fmt.Sprintf("resource-share-invitation/%s$", verify.UUIDRegexPattern))),
					resource.TestMatchResourceAttr(resourceName, "share_id", regexache.MustCompile(fmt.Sprintf(`^rs-%s$`, verify.UUIDRegexPattern))),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, string(awstypes.ResourceShareStatusActive)),
					acctest.CheckResourceAttrAccountID(resourceName, "receiver_account_id"),
					resource.TestMatchResourceAttr(resourceName, "sender_account_id", regexache.MustCompile(`\d{12}`)),
					resource.TestCheckResourceAttr(resourceName, "share_name", rName),
					resource.TestCheckResourceAttr(resourceName, "resources.%", acctest.Ct0),
				),
			},
		},
	})
}

func testAccCheckResourceShareAccepterDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).RAMClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_ram_resource_share_accepter" {
				continue
			}

			_, err := tfram.FindResourceShareOwnerOtherAccountsByARN(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("RAM Resource Share %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckResourceShareAccepterExists(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).RAMClient(ctx)

		_, err := tfram.FindResourceShareOwnerOtherAccountsByARN(ctx, conn, rs.Primary.ID)

		return err
	}
}

func testAccResourceShareAccepterConfig_basic(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigAlternateAccountProvider(), fmt.Sprintf(`
resource "aws_ram_resource_share_accepter" "test" {
  share_arn = aws_ram_principal_association.test.resource_share_arn
}

resource "aws_ram_resource_share" "test" {
  provider = "awsalternate"

  name                      = %[1]q
  allow_external_principals = true

  tags = {
    Name = %[1]q
  }
}

resource "aws_ram_principal_association" "test" {
  provider = "awsalternate"

  principal          = data.aws_caller_identity.receiver.account_id
  resource_share_arn = aws_ram_resource_share.test.arn
}

data "aws_caller_identity" "receiver" {}
`, rName))
}

func testAccResourceShareAccepterConfig_association(rName string) string {
	return acctest.ConfigCompose(testAccResourceShareAccepterConfig_basic(rName), fmt.Sprintf(`
resource "aws_ram_resource_association" "test" {
  provider = "awsalternate"

  resource_arn       = aws_codebuild_project.test.arn
  resource_share_arn = aws_ram_resource_share.test.arn
}

resource "aws_codebuild_project" "test" {
  provider = "awsalternate"

  name         = %[1]q
  service_role = aws_iam_role.test.arn

  artifacts {
    type = "NO_ARTIFACTS"
  }

  environment {
    compute_type = "BUILD_GENERAL1_SMALL"
    image        = "2"
    type         = "LINUX_CONTAINER"
  }

  source {
    type     = "GITHUB"
    location = "https://github.com/hashicorp/packer.git"
  }
}

data "aws_partition" "current" {}

resource "aws_iam_role" "test" {
  provider = "awsalternate"

  name = %[1]q

  assume_role_policy = jsonencode({
    Statement = [{
      Action = "sts:AssumeRole"
      Effect = "Allow"
      Principal = {
        Service = "codebuild.${data.aws_partition.current.dns_suffix}"
      }
    }]
    Version = "2012-10-17"
  })
}

resource "aws_iam_role_policy" "test" {
  provider = "awsalternate"

  name = %[1]q
  role = aws_iam_role.test.name

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Effect   = "Allow"
      Resource = ["*"]
      Action = [
        "logs:CreateLogGroup",
        "logs:CreateLogStream",
        "logs:PutLogEvents"
      ]
    }]
  })
}
`, rName))
}
