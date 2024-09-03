// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package signer_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/signer"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfsigner "github.com/hashicorp/terraform-provider-aws/internal/service/signer"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccSignerSigningProfilePermission_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := fmt.Sprintf("tf_acc_test_%d", sdkacctest.RandInt())
	resourceName := "aws_signer_signing_profile_permission.test_sp_permission"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckSingerSigningProfile(ctx, t, "AWSLambda-SHA384-ECDSA")
		},
		ErrorCheck:               acctest.ErrorCheck(t, signer.ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSigningProfilePermissionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config:  testAccSigningProfilePermissionConfig_basic(rName),
				Destroy: false,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSigningProfilePermissionExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrAction, "signer:StartSigningJob"),
					resource.TestCheckResourceAttr(resourceName, "profile_version", ""),
					resource.TestCheckResourceAttr(resourceName, "statement_id", rName),
					resource.TestCheckResourceAttr(resourceName, "statement_id_prefix", ""),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: testAccSigningProfilePermissionImportStateIdFunc(resourceName),
			},
		},
	})
}

func TestAccSignerSigningProfilePermission_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	rName := fmt.Sprintf("tf_acc_test_%d", sdkacctest.RandInt())
	resourceName := "aws_signer_signing_profile_permission.test_sp_permission"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckSingerSigningProfile(ctx, t, "AWSLambda-SHA384-ECDSA")
		},
		ErrorCheck:               acctest.ErrorCheck(t, signer.ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSigningProfilePermissionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config:  testAccSigningProfilePermissionConfig_basic(rName),
				Destroy: false,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSigningProfilePermissionExists(ctx, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfsigner.ResourceSigningProfilePermission(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccSignerSigningProfilePermission_statementIDGenerated(t *testing.T) {
	ctx := acctest.Context(t)
	rName := fmt.Sprintf("tf_acc_test_%d", sdkacctest.RandInt())
	resourceName := "aws_signer_signing_profile_permission.test_sp_permission"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckSingerSigningProfile(ctx, t, "AWSLambda-SHA384-ECDSA")
		},
		ErrorCheck:               acctest.ErrorCheck(t, signer.ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSigningProfilePermissionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSigningProfilePermissionConfig_statementIDGenerated(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSigningProfilePermissionExists(ctx, resourceName),
					acctest.CheckResourceAttrNameGenerated(resourceName, "statement_id"),
					resource.TestCheckResourceAttr(resourceName, "statement_id_prefix", id.UniqueIdPrefix),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: testAccSigningProfilePermissionImportStateIdFunc(resourceName),
			},
		},
	})
}

func TestAccSignerSigningProfilePermission_statementIDPrefix(t *testing.T) {
	ctx := acctest.Context(t)
	rName := fmt.Sprintf("tf_acc_test_%d", sdkacctest.RandInt())
	resourceName := "aws_signer_signing_profile_permission.test_sp_permission"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckSingerSigningProfile(ctx, t, "AWSLambda-SHA384-ECDSA")
		},
		ErrorCheck:               acctest.ErrorCheck(t, signer.ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSigningProfilePermissionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSigningProfilePermissionConfig_statementIDPrefix(rName, "tf-acc-test-prefix-"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSigningProfilePermissionExists(ctx, resourceName),
					acctest.CheckResourceAttrNameFromPrefix(resourceName, "statement_id", "tf-acc-test-prefix-"),
					resource.TestCheckResourceAttr(resourceName, "statement_id_prefix", "tf-acc-test-prefix-"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: testAccSigningProfilePermissionImportStateIdFunc(resourceName),
			},
		},
	})
}

func TestAccSignerSigningProfilePermission_getSigningProfile(t *testing.T) {
	ctx := acctest.Context(t)
	rName := fmt.Sprintf("tf_acc_test_%d", sdkacctest.RandInt())
	resourceName := "aws_signer_signing_profile_permission.test_sp_permission"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckSingerSigningProfile(ctx, t, "AWSLambda-SHA384-ECDSA")
		},
		ErrorCheck:               acctest.ErrorCheck(t, signer.ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSigningProfilePermissionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSigningProfilePermissionConfig_getSP(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSigningProfilePermissionExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrAction, "signer:GetSigningProfile"),
				),
			},
			{
				Config: testAccSigningProfilePermissionConfig_revokeSignature(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSigningProfilePermissionExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrAction, "signer:RevokeSignature"),
				),
			},
		},
	})
}

func TestAccSignerSigningProfilePermission_StartSigningJob_getSP(t *testing.T) {
	ctx := acctest.Context(t)
	rName := fmt.Sprintf("tf_acc_test_%d", sdkacctest.RandInt())
	resourceName1 := "aws_signer_signing_profile_permission.sp1_perm"
	resourceName2 := "aws_signer_signing_profile_permission.sp2_perm"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckSingerSigningProfile(ctx, t, "AWSLambda-SHA384-ECDSA")
		},
		ErrorCheck:               acctest.ErrorCheck(t, signer.ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSigningProfilePermissionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSigningProfilePermissionConfig_startJobGetSP(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSigningProfilePermissionExists(ctx, resourceName1),
					testAccCheckSigningProfilePermissionExists(ctx, resourceName2),
				),
			},
		},
	})
}

func testAccCheckSigningProfilePermissionExists(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).SignerClient(ctx)

		_, err := tfsigner.FindPermissionByTwoPartKey(ctx, conn, rs.Primary.Attributes["profile_name"], rs.Primary.Attributes["statement_id"])

		return err
	}
}

func testAccCheckSigningProfilePermissionDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).SignerClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_signer_signing_profile_permission" {
				continue
			}

			_, err := tfsigner.FindPermissionByTwoPartKey(ctx, conn, rs.Primary.Attributes["profile_name"], rs.Primary.Attributes["statement_id"])

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Signer Signing Profile Permission %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccSigningProfilePermissionImportStateIdFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("Not found: %s", resourceName)
		}

		return fmt.Sprintf("%s/%s", rs.Primary.Attributes["profile_name"], rs.Primary.Attributes["statement_id"]), nil
	}
}

func testAccSigningProfilePermissionConfig_base(rName string) string {
	return fmt.Sprintf(`
resource "aws_signer_signing_profile" "test_sp" {
  platform_id = "AWSLambda-SHA384-ECDSA"
  name        = %[1]q
}`, rName)
}

func testAccSigningProfilePermissionConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccSigningProfilePermissionConfig_base(rName), fmt.Sprintf(`
data "aws_caller_identity" "current" {}

resource "aws_signer_signing_profile_permission" "test_sp_permission" {
  profile_name = aws_signer_signing_profile.test_sp.name
  action       = "signer:StartSigningJob"
  principal    = data.aws_caller_identity.current.account_id
  statement_id = %[1]q
}`, rName))
}

func testAccSigningProfilePermissionConfig_statementIDGenerated(rName string) string {
	return acctest.ConfigCompose(testAccSigningProfilePermissionConfig_base(rName), `
data "aws_caller_identity" "current" {}

resource "aws_signer_signing_profile_permission" "test_sp_permission" {
  profile_name = aws_signer_signing_profile.test_sp.name
  action       = "signer:StartSigningJob"
  principal    = data.aws_caller_identity.current.account_id
}`)
}

func testAccSigningProfilePermissionConfig_statementIDPrefix(rName, prefix string) string {
	return acctest.ConfigCompose(testAccSigningProfilePermissionConfig_base(rName), fmt.Sprintf(`
data "aws_caller_identity" "current" {}

resource "aws_signer_signing_profile_permission" "test_sp_permission" {
  profile_name        = aws_signer_signing_profile.test_sp.name
  action              = "signer:StartSigningJob"
  principal           = data.aws_caller_identity.current.account_id
  statement_id_prefix = %[1]q
}`, prefix))
}

func testAccSigningProfilePermissionConfig_startJobGetSP(rName string) string {
	return acctest.ConfigCompose(testAccSigningProfilePermissionConfig_base(rName), `
data "aws_caller_identity" "current" {}

resource "aws_signer_signing_profile_permission" "sp1_perm" {
  profile_name = aws_signer_signing_profile.test_sp.name
  action       = "signer:StartSigningJob"
  principal    = data.aws_caller_identity.current.account_id
  statement_id = "statementid1"
}

resource "aws_signer_signing_profile_permission" "sp2_perm" {
  profile_name = aws_signer_signing_profile.test_sp.name
  action       = "signer:GetSigningProfile"
  principal    = data.aws_caller_identity.current.account_id
  statement_id = "statementid2"
}`)
}

func testAccSigningProfilePermissionConfig_getSP(rName string) string {
	return acctest.ConfigCompose(testAccSigningProfilePermissionConfig_base(rName), fmt.Sprintf(`
data "aws_caller_identity" "current" {}

resource "aws_signer_signing_profile_permission" "test_sp_permission" {
  profile_name = aws_signer_signing_profile.test_sp.name
  action       = "signer:GetSigningProfile"
  principal    = data.aws_caller_identity.current.account_id
  statement_id = %[1]q
}`, rName))
}

func testAccSigningProfilePermissionConfig_revokeSignature(rName string) string {
	return acctest.ConfigCompose(testAccSigningProfilePermissionConfig_base(rName), fmt.Sprintf(`
data "aws_caller_identity" "current" {}

resource "aws_signer_signing_profile_permission" "test_sp_permission" {
  profile_name = aws_signer_signing_profile.test_sp.name
  action       = "signer:RevokeSignature"
  principal    = data.aws_caller_identity.current.account_id
  statement_id = %[1]q
}`, rName))
}
