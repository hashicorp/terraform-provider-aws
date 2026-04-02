// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package iam_test

import (
	"context"
	"fmt"
	"testing"

	awstypes "github.com/aws/aws-sdk-go-v2/service/iam/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfiam "github.com/hashicorp/terraform-provider-aws/internal/service/iam"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccIAMServiceSpecificCredential_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var cred awstypes.ServiceSpecificCredentialMetadata

	resourceName := "aws_iam_service_specific_credential.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IAMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckServiceSpecificCredentialDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccServiceSpecificCredentialConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceSpecificCredentialExists(ctx, t, resourceName, &cred),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrUserName, "aws_iam_user.test", names.AttrName),
					resource.TestCheckResourceAttr(resourceName, names.AttrServiceName, "codecommit.amazonaws.com"),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, "Active"),
					resource.TestCheckResourceAttrSet(resourceName, "service_user_name"),
					resource.TestCheckResourceAttrSet(resourceName, "service_specific_credential_id"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"service_password", "service_credential_secret"},
			},
		},
	})
}

func TestAccIAMServiceSpecificCredential_multi(t *testing.T) {
	ctx := acctest.Context(t)
	var cred awstypes.ServiceSpecificCredentialMetadata

	resourceName := "aws_iam_service_specific_credential.test"
	resourceName2 := "aws_iam_service_specific_credential.test2"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IAMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckServiceSpecificCredentialDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccServiceSpecificCredentialConfig_multi(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceSpecificCredentialExists(ctx, t, resourceName, &cred),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrUserName, "aws_iam_user.test", names.AttrName),
					resource.TestCheckResourceAttr(resourceName, names.AttrServiceName, "codecommit.amazonaws.com"),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, "Active"),
					resource.TestCheckResourceAttrSet(resourceName, "service_user_name"),
					resource.TestCheckResourceAttrSet(resourceName, "service_specific_credential_id"),
					resource.TestCheckResourceAttrPair(resourceName2, names.AttrUserName, "aws_iam_user.test", names.AttrName),
					resource.TestCheckResourceAttr(resourceName2, names.AttrServiceName, "codecommit.amazonaws.com"),
					resource.TestCheckResourceAttr(resourceName2, names.AttrStatus, "Active"),
					resource.TestCheckResourceAttrSet(resourceName2, "service_user_name"),
					resource.TestCheckResourceAttrSet(resourceName2, "service_specific_credential_id"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"service_password", "service_credential_secret"},
			},
		},
	})
}

func TestAccIAMServiceSpecificCredential_status(t *testing.T) {
	ctx := acctest.Context(t)
	var cred awstypes.ServiceSpecificCredentialMetadata

	resourceName := "aws_iam_service_specific_credential.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IAMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckServiceSpecificCredentialDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccServiceSpecificCredentialConfig_status(rName, "Inactive"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceSpecificCredentialExists(ctx, t, resourceName, &cred),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, "Inactive"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"service_password", "service_credential_secret"},
			},
			{
				Config: testAccServiceSpecificCredentialConfig_status(rName, "Active"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceSpecificCredentialExists(ctx, t, resourceName, &cred),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, "Active"),
				),
			},
			{
				Config: testAccServiceSpecificCredentialConfig_status(rName, "Inactive"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceSpecificCredentialExists(ctx, t, resourceName, &cred),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, "Inactive"),
				),
			},
		},
	})
}

func TestAccIAMServiceSpecificCredential_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var cred awstypes.ServiceSpecificCredentialMetadata
	resourceName := "aws_iam_service_specific_credential.test"

	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IAMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckServiceSpecificCredentialDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccServiceSpecificCredentialConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceSpecificCredentialExists(ctx, t, resourceName, &cred),
					acctest.CheckSDKResourceDisappears(ctx, t, tfiam.ResourceServiceSpecificCredential(), resourceName),
					acctest.CheckSDKResourceDisappears(ctx, t, tfiam.ResourceServiceSpecificCredential(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckServiceSpecificCredentialExists(ctx context.Context, t *testing.T, n string, v *awstypes.ServiceSpecificCredentialMetadata) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).IAMClient(ctx)

		output, err := tfiam.FindServiceSpecificCredentialByThreePartKey(ctx, conn, rs.Primary.Attributes[names.AttrServiceName], rs.Primary.Attributes[names.AttrUserName], rs.Primary.Attributes["service_specific_credential_id"])

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckServiceSpecificCredentialDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).IAMClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_iam_service_specific_credential" {
				continue
			}

			output, err := tfiam.FindServiceSpecificCredentialByThreePartKey(ctx, conn, rs.Primary.Attributes[names.AttrServiceName], rs.Primary.Attributes[names.AttrUserName], rs.Primary.Attributes["service_specific_credential_id"])

			if retry.NotFound(err) {
				continue
			}

			if output != nil {
				return fmt.Errorf("IAM Service-Specific Credential (%s) still exists", rs.Primary.ID)
			}
		}

		return nil
	}
}

func testAccServiceSpecificCredentialConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_iam_user" "test" {
  name = %[1]q
}

resource "aws_iam_service_specific_credential" "test" {
  service_name = "codecommit.amazonaws.com"
  user_name    = aws_iam_user.test.name
}
`, rName)
}

func testAccServiceSpecificCredentialConfig_multi(rName string) string {
	return fmt.Sprintf(`
resource "aws_iam_user" "test" {
  name = %[1]q
}

resource "aws_iam_service_specific_credential" "test" {
  service_name = "codecommit.amazonaws.com"
  user_name    = aws_iam_user.test.name
}

resource "aws_iam_service_specific_credential" "test2" {
  service_name = "codecommit.amazonaws.com"
  user_name    = aws_iam_user.test.name
}
`, rName)
}

func testAccServiceSpecificCredentialConfig_status(rName, status string) string {
	return fmt.Sprintf(`
resource "aws_iam_user" "test" {
  name = %[1]q
}

resource "aws_iam_service_specific_credential" "test" {
  service_name = "codecommit.amazonaws.com"
  user_name    = aws_iam_user.test.name
  status       = %[2]q
}
`, rName, status)
}

func TestAccIAMServiceSpecificCredential_bedrockWithExpiration(t *testing.T) {
	ctx := acctest.Context(t)
	var cred awstypes.ServiceSpecificCredentialMetadata

	resourceName := "aws_iam_service_specific_credential.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IAMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckServiceSpecificCredentialDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccServiceSpecificCredentialConfig_bedrockWithExpiration(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceSpecificCredentialExists(ctx, t, resourceName, &cred),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrUserName, "aws_iam_user.test", names.AttrName),
					resource.TestCheckResourceAttr(resourceName, names.AttrServiceName, "bedrock.amazonaws.com"),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, "Active"),
					resource.TestCheckResourceAttr(resourceName, "credential_age_days", "30"),
					resource.TestCheckResourceAttrSet(resourceName, "service_credential_alias"),
					resource.TestCheckResourceAttrSet(resourceName, "service_specific_credential_id"),
					resource.TestCheckResourceAttrSet(resourceName, "create_date"),
					resource.TestCheckResourceAttrSet(resourceName, "expiration_date"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"service_password", "service_credential_secret", "credential_age_days"},
			},
		},
	})
}

func testAccServiceSpecificCredentialConfig_bedrockWithExpiration(rName string) string {
	return fmt.Sprintf(`
resource "aws_iam_user" "test" {
  name = %[1]q
}

resource "aws_iam_service_specific_credential" "test" {
  service_name        = "bedrock.amazonaws.com"
  user_name           = aws_iam_user.test.name
  credential_age_days = 30
}
`, rName)
}
