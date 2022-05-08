package signer_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/signer"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
)

func TestAccSignerSigningProfilePermission_basic(t *testing.T) {
	resourceName := "aws_signer_signing_profile_permission.test_sp_permission"
	profileResourceName := "aws_signer_signing_profile.test_sp"
	rString := sdkacctest.RandString(53)
	profileName := fmt.Sprintf("tf_acc_spp_%s", rString)

	var conf signer.GetSigningProfileOutput
	var sppconf signer.ListProfilePermissionsOutput

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckSingerSigningProfile(t, "AWSLambda-SHA384-ECDSA") },
		ErrorCheck:        acctest.ErrorCheck(t, signer.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckSigningProfileDestroy,
		Steps: []resource.TestStep{
			{
				Config:  testAccSigningProfilePermissionConfig(profileName),
				Destroy: false,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSigningProfileExists(profileResourceName, &conf),
					testAccCheckSigningProfilePermissionExists(resourceName, profileName, &sppconf),
					create.TestCheckResourceAttrNameGenerated(resourceName, "statement_id"),
				),
			},
			{
				ResourceName:            profileResourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"name_prefix"},
			},
		},
	})
}

func TestAccSignerSigningProfilePermission_getSigningProfile(t *testing.T) {
	resourceName := "aws_signer_signing_profile_permission.test_sp_permission"
	profileResourceName := "aws_signer_signing_profile.test_sp"
	rString := sdkacctest.RandString(53)
	profileName := fmt.Sprintf("tf_acc_spp_%s", rString)

	var conf signer.GetSigningProfileOutput
	var sppconf signer.ListProfilePermissionsOutput

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckSingerSigningProfile(t, "AWSLambda-SHA384-ECDSA") },
		ErrorCheck:        acctest.ErrorCheck(t, signer.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckSigningProfileDestroy,
		Steps: []resource.TestStep{
			{
				Config:  testAccSigningProfilePermissionGetSP(profileName),
				Destroy: false,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSigningProfileExists(profileResourceName, &conf),
					testAccCheckSigningProfilePermissionExists(resourceName, profileName, &sppconf),
				),
			},
			{
				ResourceName:            profileResourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"name_prefix"},
			},
			{
				Config:  testAccSigningProfilePermissionRevokeSignature(profileName),
				Destroy: false,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSigningProfileExists(profileResourceName, &conf),
					testAccCheckSigningProfilePermissionExists(resourceName, profileName, &sppconf),
				),
			},
		},
	})
}

func TestAccSignerSigningProfilePermission_StartSigningJob_getSP(t *testing.T) {
	resourceName1 := "aws_signer_signing_profile_permission.sp1_perm"
	resourceName2 := "aws_signer_signing_profile_permission.sp2_perm"
	profileResourceName := "aws_signer_signing_profile.test_sp"
	rString := sdkacctest.RandString(53)
	profileName := fmt.Sprintf("tf_acc_spp_%s", rString)

	var conf signer.GetSigningProfileOutput
	var sppconf signer.ListProfilePermissionsOutput

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckSingerSigningProfile(t, "AWSLambda-SHA384-ECDSA") },
		ErrorCheck:        acctest.ErrorCheck(t, signer.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckSigningProfileDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSigningProfilePermissionStartSigningJobGetSP(profileName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSigningProfileExists(profileResourceName, &conf),
					testAccCheckSigningProfilePermissionExists(resourceName1, profileName, &sppconf),
					testAccCheckSigningProfilePermissionExists(resourceName2, profileName, &sppconf),
				),
			},
			{
				ResourceName:            profileResourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"name_prefix"},
			},
		},
	})
}

func TestAccSignerSigningProfilePermission_statementPrefix(t *testing.T) {
	resourceName := "aws_signer_signing_profile_permission.sp1_perm"
	profileResourceName := "aws_signer_signing_profile.test_sp"
	rString := sdkacctest.RandString(53)
	profileName := fmt.Sprintf("tf_acc_spp_%s", rString)
	statementNamePrefix := "tf_acc_spp_statement_"

	//var conf signer.GetSigningProfileOutput
	var sppconf signer.ListProfilePermissionsOutput

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckSingerSigningProfile(t, "AWSLambda-SHA384-ECDSA") },
		ErrorCheck:        acctest.ErrorCheck(t, signer.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckSigningProfileDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSigningProfilePermissionStatementPrefix(statementNamePrefix, profileName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSigningProfilePermissionExists(resourceName, profileName, &sppconf),
					create.TestCheckResourceAttrNameFromPrefix(resourceName, "statement_id", statementNamePrefix),
				),
			},
			{
				ResourceName:            profileResourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"name_prefix"},
			},
		},
	})
}

func testAccSigningProfilePermissionConfig(profileName string) string {
	return fmt.Sprintf(testAccSigningProfilePermissionConfig_base(profileName) + `
data "aws_caller_identity" "current" {}

resource "aws_signer_signing_profile_permission" "test_sp_permission" {
  profile_name = aws_signer_signing_profile.test_sp.name
  action       = "signer:StartSigningJob"
  principal    = data.aws_caller_identity.current.account_id
}`)
}

func testAccSigningProfilePermissionStartSigningJobGetSP(profileName string) string {
	return fmt.Sprintf(testAccSigningProfilePermissionConfig_base(profileName) + `
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

func testAccSigningProfilePermissionStatementPrefix(statementNamePrefix, profileName string) string {
	return fmt.Sprintf(testAccSigningProfilePermissionConfig_base(profileName)+`
data "aws_caller_identity" "current" {}

resource "aws_signer_signing_profile_permission" "sp1_perm" {
  profile_name        = aws_signer_signing_profile.test_sp.name
  action              = "signer:StartSigningJob"
  principal           = data.aws_caller_identity.current.account_id
  statement_id_prefix = %[1]q
}`, statementNamePrefix)
}

func testAccSigningProfilePermissionGetSP(profileName string) string {
	return fmt.Sprintf(testAccSigningProfilePermissionConfig_base(profileName) + `
data "aws_caller_identity" "current" {}

resource "aws_signer_signing_profile_permission" "test_sp_permission" {
  profile_name = aws_signer_signing_profile.test_sp.name
  action       = "signer:GetSigningProfile"
  principal    = data.aws_caller_identity.current.account_id
}`)
}

func testAccSigningProfilePermissionRevokeSignature(profileName string) string {
	return fmt.Sprintf(testAccSigningProfilePermissionConfig_base(profileName) + `
data "aws_caller_identity" "current" {}

resource "aws_signer_signing_profile_permission" "test_sp_permission" {
  profile_name = aws_signer_signing_profile.test_sp.name
  action       = "signer:RevokeSignature"
  principal    = data.aws_caller_identity.current.account_id
}`)
}

func testAccSigningProfilePermissionConfig_base(profileName string) string {
	return fmt.Sprintf(`
resource "aws_signer_signing_profile" "test_sp" {
  platform_id = "AWSLambda-SHA384-ECDSA"
  name        = "%s"
}`, profileName)
}

func testAccCheckSigningProfilePermissionExists(res, profileName string, spp *signer.ListProfilePermissionsOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[res]
		if !ok {
			return fmt.Errorf("Signing profile permission not found: %s", res)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("Signing Profile with that ARN does not exist")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).SignerConn

		params := &signer.ListProfilePermissionsInput{
			ProfileName: aws.String(profileName),
		}

		getSp, err := conn.ListProfilePermissions(params)
		if err != nil {
			return err
		}

		*spp = *getSp

		return nil
	}
}
