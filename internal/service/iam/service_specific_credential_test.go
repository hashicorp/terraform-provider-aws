// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package iam_test

import (
	"context"
	"fmt"
	"testing"

	awstypes "github.com/aws/aws-sdk-go-v2/service/iam/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfiam "github.com/hashicorp/terraform-provider-aws/internal/service/iam"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccIAMServiceSpecificCredential_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var cred awstypes.ServiceSpecificCredentialMetadata

	resourceName := "aws_iam_service_specific_credential.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IAMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckServiceSpecificCredentialDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccServiceSpecificCredentialConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceSpecificCredentialExists(ctx, resourceName, &cred),
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
				ImportStateVerifyIgnore: []string{"service_password"},
			},
		},
	})
}

func TestAccIAMServiceSpecificCredential_multi(t *testing.T) {
	ctx := acctest.Context(t)
	var cred awstypes.ServiceSpecificCredentialMetadata

	resourceName := "aws_iam_service_specific_credential.test"
	resourceName2 := "aws_iam_service_specific_credential.test2"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IAMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckServiceSpecificCredentialDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccServiceSpecificCredentialConfig_multi(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceSpecificCredentialExists(ctx, resourceName, &cred),
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
				ImportStateVerifyIgnore: []string{"service_password"},
			},
		},
	})
}

func TestAccIAMServiceSpecificCredential_status(t *testing.T) {
	ctx := acctest.Context(t)
	var cred awstypes.ServiceSpecificCredentialMetadata

	resourceName := "aws_iam_service_specific_credential.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IAMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckServiceSpecificCredentialDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccServiceSpecificCredentialConfig_status(rName, "Inactive"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceSpecificCredentialExists(ctx, resourceName, &cred),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, "Inactive"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"service_password"},
			},
			{
				Config: testAccServiceSpecificCredentialConfig_status(rName, "Active"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceSpecificCredentialExists(ctx, resourceName, &cred),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, "Active"),
				),
			},
			{
				Config: testAccServiceSpecificCredentialConfig_status(rName, "Inactive"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceSpecificCredentialExists(ctx, resourceName, &cred),
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

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IAMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckServiceSpecificCredentialDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccServiceSpecificCredentialConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceSpecificCredentialExists(ctx, resourceName, &cred),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfiam.ResourceServiceSpecificCredential(), resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfiam.ResourceServiceSpecificCredential(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckServiceSpecificCredentialExists(ctx context.Context, n string, cred *awstypes.ServiceSpecificCredentialMetadata) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Server Cert ID is set")
		}
		conn := acctest.Provider.Meta().(*conns.AWSClient).IAMClient(ctx)

		serviceName, userName, credId, err := tfiam.DecodeServiceSpecificCredentialId(rs.Primary.ID)
		if err != nil {
			return err
		}

		output, err := tfiam.FindServiceSpecificCredential(ctx, conn, serviceName, userName, credId)
		if err != nil {
			return err
		}

		*cred = *output

		return nil
	}
}

func testAccCheckServiceSpecificCredentialDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).IAMClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_iam_service_specific_credential" {
				continue
			}

			serviceName, userName, credId, err := tfiam.DecodeServiceSpecificCredentialId(rs.Primary.ID)
			if err != nil {
				return err
			}

			output, err := tfiam.FindServiceSpecificCredential(ctx, conn, serviceName, userName, credId)

			if tfresource.NotFound(err) {
				continue
			}

			if output != nil {
				return fmt.Errorf("IAM Service Specific Credential (%s) still exists", rs.Primary.ID)
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
