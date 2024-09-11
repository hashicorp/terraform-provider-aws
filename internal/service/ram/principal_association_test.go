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
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccRAMPrincipalAssociation_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var association awstypes.ResourceShareAssociation
	resourceName := "aws_ram_principal_association.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckSharingWithOrganizationEnabled(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.RAMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPrincipalAssociationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPrincipalAssociationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPrincipalAssociationExists(ctx, resourceName, &association),
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

func TestAccRAMPrincipalAssociation_AccountID(t *testing.T) {
	ctx := acctest.Context(t)
	var association awstypes.ResourceShareAssociation
	resourceName := "aws_ram_principal_association.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckAlternateAccount(t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.RAMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesAlternate(ctx, t),
		CheckDestroy:             testAccCheckPrincipalAssociationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPrincipalAssociationConfig_accountID(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPrincipalAssociationExists(ctx, resourceName, &association),
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

func TestAccRAMPrincipalAssociation_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var association awstypes.ResourceShareAssociation
	resourceName := "aws_ram_principal_association.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckSharingWithOrganizationEnabled(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.RAMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPrincipalAssociationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPrincipalAssociationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPrincipalAssociationExists(ctx, resourceName, &association),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfram.ResourcePrincipalAssociation(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccRAMPrincipalAssociation_duplicate(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckSharingWithOrganizationEnabled(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.RAMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPrincipalAssociationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config:      testAccPrincipalAssociationConfig_duplicate(rName),
				ExpectError: regexache.MustCompile(`RAM Principal Association .* already exists`),
			},
		},
	})
}

func testAccPreCheckSharingWithOrganizationEnabled(ctx context.Context, t *testing.T) {
	err := tfram.FindSharingWithOrganization(ctx, acctest.Provider.Meta().(*conns.AWSClient))

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if tfresource.NotFound(err) {
		t.Skipf("Sharing with AWS Organization not found, skipping acceptance test: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccCheckPrincipalAssociationExists(ctx context.Context, n string, v *awstypes.ResourceShareAssociation) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).RAMClient(ctx)

		output, err := tfram.FindPrincipalAssociationByTwoPartKey(ctx, conn, rs.Primary.Attributes["resource_share_arn"], rs.Primary.Attributes[names.AttrPrincipal])

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckPrincipalAssociationDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).RAMClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_ram_principal_association" {
				continue
			}

			_, err := tfram.FindPrincipalAssociationByTwoPartKey(ctx, conn, rs.Primary.Attributes["resource_share_arn"], rs.Primary.Attributes[names.AttrPrincipal])

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("RAM Principal Association %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccPrincipalAssociationConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_ram_resource_share" "test" {
  allow_external_principals = false
  name                      = %[1]q
}

data "aws_partition" "current" {}

resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Action = "sts:AssumeRole",
      Principal = {
        Service = "ec2.${data.aws_partition.current.dns_suffix}",
      }
      Effect = "Allow"
    }]
  })
}

resource "aws_ram_principal_association" "test" {
  principal          = aws_iam_role.test.arn
  resource_share_arn = aws_ram_resource_share.test.id
}
`, rName)
}

func testAccPrincipalAssociationConfig_accountID(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigAlternateAccountProvider(), fmt.Sprintf(`
resource "aws_ram_resource_share" "test" {
  allow_external_principals = true
  name                      = %[1]q
}

data "aws_caller_identity" "receiver" {
  provider = "awsalternate"
}

resource "aws_ram_principal_association" "test" {
  principal          = data.aws_caller_identity.receiver.account_id
  resource_share_arn = aws_ram_resource_share.test.id
}
`, rName))
}

func testAccPrincipalAssociationConfig_duplicate(rName string) string {
	return fmt.Sprintf(`
resource "aws_ram_resource_share" "test" {
  allow_external_principals = false
  name                      = %[1]q
}

data "aws_partition" "current" {}

resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Action = "sts:AssumeRole",
      Principal = {
        Service = "ec2.${data.aws_partition.current.dns_suffix}",
      }
      Effect = "Allow"
    }]
  })
}

resource "aws_ram_principal_association" "test1" {
  principal          = aws_iam_role.test.arn
  resource_share_arn = aws_ram_resource_share.test.id
}

resource "aws_ram_principal_association" "test2" {
  principal          = aws_iam_role.test.arn
  resource_share_arn = aws_ram_principal_association.test1.resource_share_arn
}
`, rName)
}
